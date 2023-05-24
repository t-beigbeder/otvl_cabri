package cabridss

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type IS3Session interface {
	Initialize() error
	Check() error
	List(prefix string) ([]string, error)
	Put(key string, content []byte) error
	Get(key string) ([]byte, error)
	Upload(key string, r io.Reader) error
	Download(key string) (io.ReadCloser, error)
	Delete(key string) error
	DeleteAll(prefix string) error
}

func (s3s *s3Session) Initialize() error {
	if s3s.getMock != nil {
		s3s.mock = s3s.getMock(s3s)
		if err := s3s.mock.Initialize(); err != nil {
			return err
		}
	}
	epr := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		return endpoints.ResolvedEndpoint{
			URL: s3s.config.Endpoint,
		}, nil
	}
	config := aws.Config{
		Credentials:      credentials.NewStaticCredentials(s3s.config.AccessKey, s3s.config.SecretKey, ""),
		Region:           aws.String(s3s.config.Region),
		EndpointResolver: endpoints.ResolverFunc(epr),
	}
	s3s.session = session.Must(session.NewSession(&config))
	s3s.s3Svc = s3.New(s3s.session)
	return nil
}

func (s3s *s3Session) Check() error {
	_, err := s3s.s3Svc.HeadBucket(&s3.HeadBucketInput{Bucket: aws.String(s3s.config.Container)})
	return err
}

func (s3s *s3Session) List(prefix string) ([]string, error) {
	var res []string
	err := s3s.s3Svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(s3s.config.Container),
		Prefix: aws.String(prefix),
	}, func(output *s3.ListObjectsV2Output, b bool) bool {
		for i := 0; i < int(*output.KeyCount); i++ {
			res = append(res, *output.Contents[i].Key)
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	if s3s.mock != nil {
		if _, err = s3s.mock.List(prefix); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (s3s *s3Session) Put(key string, content []byte) error {
	if _, err := s3s.s3Svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3s.config.Container),
		Key:    aws.String(key),
		Body:   aws.ReadSeekCloser(strings.NewReader(string(content))),
	}); err != nil {
		return err
	}
	if s3s.mock != nil {
		return s3s.mock.Put(key, content)
	}
	return nil
}

func (s3s *s3Session) Get(key string) ([]byte, error) {
	res, err := s3s.s3Svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3s.config.Container),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var b bytes.Buffer
	_, err = io.CopyN(&b, res.Body, MAX_META_SIZE)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if len(b.Bytes()) == MAX_META_SIZE {
		return nil, fmt.Errorf("meta size exceeded %d", MAX_META_SIZE)
	}
	if s3s.mock != nil {
		if _, err = s3s.mock.Get(key); err != nil {
			return nil, err
		}
	}
	return b.Bytes(), nil
}

func (s3s *s3Session) Upload(key string, r io.Reader) error {
	var buf bytes.Buffer
	if s3s.mock != nil {
		r = io.TeeReader(r, &buf)
	}
	uploader := s3manager.NewUploader(s3s.session)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3s.config.Container),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return err
	}
	if s3s.mock != nil {
		return s3s.mock.Upload(key, &buf)
	}
	return nil
}

func (s3s *s3Session) Download(key string) (io.ReadCloser, error) {
	res, err := s3s.s3Svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3s.config.Container),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	if s3s.mock != nil {
		rc, err := s3s.mock.Download(key)
		if err != nil {
			res.Body.Close()
			return nil, err
		}
		if rc != nil {
			rc.Close()
		}
	}
	return res.Body, nil
}

func (s3s *s3Session) Delete(key string) error {
	if _, err := s3s.s3Svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s3s.config.Container),
		Key:    aws.String(key),
	}); err != nil {
		return err
	}
	if s3s.mock != nil {
		return s3s.mock.Delete(key)
	}
	return nil
}

func (s3s *s3Session) DeleteAll(prefix string) error {
	cs, err := s3s.List(prefix)
	if err != nil {
		return err
	}

	MAX := 32
	var wg sync.WaitGroup
	var mx sync.Mutex
	var lerr error
	ch := make(chan string)
	for i := 0; i < MAX; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				c := <-ch
				if c == "" {
					return
				}
				if err = s3s.Delete(c); err != nil {
					time.Sleep(100 * time.Millisecond)
					if err = s3s.Delete(c); err != nil {
						mx.Lock()
						lerr = err
						mx.Unlock()
					}
				}
			}
		}()
	}

	for _, c := range cs {
		ch <- c
	}
	close(ch)
	wg.Wait()
	if lerr != nil {
		return lerr
	}
	if s3s.mock != nil {
		return s3s.mock.DeleteAll(prefix)
	}
	return nil
}

type s3Session struct {
	getMock func(IS3Session) IS3Session
	mock    IS3Session
	config  ObsConfig
	session *session.Session
	s3Svc   *s3.S3
}

func NewS3Session(config ObsConfig, getMock func(IS3Session) IS3Session) IS3Session {
	return &s3Session{config: config, getMock: getMock}
}

func CleanS3Session(config ObsConfig, getMock func(IS3Session) IS3Session) (IS3Session, error) {
	s3s := &s3Session{config: config, getMock: getMock}
	err := s3s.DeleteAll("")
	return s3s, err
}

func (s3m *s3sMockFs) Initialize() error {
	if s3m.getMock != nil {
		s3m.mock = s3m.getMock(s3m)
		if err := s3m.mock.Initialize(); err != nil {
			return err
		}
	}
	return nil
}

func (s3m *s3sMockFs) List(prefix string) ([]string, error) {
	s3m.lock.Lock()
	defer s3m.lock.Unlock()
	if len(s3m.cache) == 0 {
		tfi, err := os.ReadDir(ufpath.Join(s3m.root, ufpath.Dir(prefix)))
		if err != nil {
			return nil, fmt.Errorf("in List: %w", err)
		}
		for _, fi := range tfi {
			s3m.cache[fi.Name()] = true
		}
	}

	var children []string
	fPrefix := ufpath.Base(prefix)
	for nm, _ := range s3m.cache {
		if strings.HasPrefix(nm, fPrefix) {
			children = append(children, nm)
		}
	}
	if s3m.mock != nil {
		if _, err := s3m.mock.List(prefix); err != nil {
			return nil, err
		}
	}
	return children, nil
}

func (s3m *s3sMockFs) Put(key string, content []byte) error {
	s3m.lock.Lock()
	defer s3m.lock.Unlock()

	fo, err := os.Create(ufpath.Join(s3m.root, key))
	if err != nil {
		return err
	}
	defer fo.Close()
	s3m.cache[key] = true
	_, err = fo.Write(content)
	if err != nil {
		return err
	}
	err = fo.Close()
	if err != nil {
		return err
	}
	if s3m.mock != nil {
		return s3m.mock.Put(key, content)
	}
	return nil
}

func (s3m *s3sMockFs) Get(key string) ([]byte, error) {
	fi, err := os.Open(ufpath.Join(s3m.root, key))
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	var b bytes.Buffer
	_, err = io.CopyN(&b, fi, MAX_META_SIZE)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if len(b.Bytes()) == MAX_META_SIZE {
		return nil, fmt.Errorf("meta size exceeded %d", MAX_META_SIZE)
	}
	if s3m.mock != nil {
		if _, err = s3m.mock.Get(key); err != nil {
			return nil, err
		}
	}
	return b.Bytes(), nil
}

func (s3m *s3sMockFs) Upload(key string, r io.Reader) error {
	s3m.lock.Lock()
	defer s3m.lock.Unlock()

	var buf bytes.Buffer
	if s3m.mock != nil {
		r = io.TeeReader(r, &buf)
	}
	fo, err := os.Create(ufpath.Join(s3m.root, key))
	if err != nil {
		return err
	}
	defer fo.Close()
	s3m.cache[key] = true
	if _, err = io.Copy(fo, r); err != nil {
		return err
	}
	err = fo.Close()
	if err != nil {
		return err
	}
	if s3m.mock != nil {
		return s3m.mock.Upload(key, &buf)
	}
	return nil
}

func (s3m *s3sMockFs) Download(key string) (io.ReadCloser, error) {
	rc, err := os.Open(ufpath.Join(s3m.root, key))
	if err != nil {
		return nil, err
	}
	if s3m.mock != nil {
		mrc, err := s3m.mock.Download(key)
		if err != nil {
			rc.Close()
			return nil, err
		}
		if mrc != nil {
			mrc.Close()
		}
	}

	return rc, nil
}

func (s3m *s3sMockFs) Delete(key string) error {
	s3m.lock.Lock()
	defer s3m.lock.Unlock()

	if err := os.Remove(ufpath.Join(s3m.root, key)); err != nil {
		return err
	}
	delete(s3m.cache, key)
	if s3m.mock != nil {
		return s3m.mock.Delete(key)
	}
	return nil
}

func (s3m *s3sMockFs) DeleteAll(prefix string) error {
	cs, err := s3m.List(prefix)
	if err != nil {
		return err
	}

	s3m.lock.Lock()
	defer s3m.lock.Unlock()
	s3m.cache = map[string]bool{}

	var lerr error
	for _, c := range cs {
		if err = os.Remove(ufpath.Join(s3m.root, c)); err != nil {
			lerr = err
		}
	}
	if lerr != nil {
		return lerr
	}
	if s3m.mock != nil {
		return s3m.mock.DeleteAll(prefix)
	}
	return nil
}

type s3sMockFs struct {
	getMock func(IS3Session) IS3Session
	root    string
	mock    IS3Session
	cache   map[string]bool
	lock    sync.Mutex
}

func (s3m *s3sMockFs) Check() error { return nil }

func NewS3sMockFs(root string, getMock func(IS3Session) IS3Session) IS3Session {
	return &s3sMockFs{root: root, getMock: getMock, cache: map[string]bool{}}
}

func (s3t s3sMockTests) Initialize() error {
	if s3t.testsCb != nil {
		if err := s3t.testsCb(s3t, "Initialize"); err != nil {
			return err.(error)
		}
	}
	return nil
}

func (s3t s3sMockTests) List(prefix string) ([]string, error) {
	if err := s3t.testsCb(s3t, "List", prefix); err != nil {
		return nil, err.(error)
	}
	return nil, nil
}

func (s3t s3sMockTests) Put(key string, content []byte) error {
	if err := s3t.testsCb(s3t, "Put", key, content); err != nil {
		return err.(error)
	}
	return nil
}

func (s3t s3sMockTests) Get(key string) ([]byte, error) {
	if err := s3t.testsCb(s3t, "Get", key); err != nil {
		return nil, err.(error)
	}
	return nil, nil
}

func (s3t s3sMockTests) Upload(key string, r io.Reader) error {
	if err := s3t.testsCb(s3t, "Upload", key, r); err != nil {
		return err.(error)
	}
	return nil
}

func (s3t s3sMockTests) Download(key string) (io.ReadCloser, error) {
	if err := s3t.testsCb(s3t, "Download", key); err != nil {
		return nil, err.(error)
	}
	return nil, nil
}

func (s3t s3sMockTests) Delete(key string) error {
	if err := s3t.testsCb(s3t, "Delete", key); err != nil {
		return err.(error)
	}
	return nil
}

func (s3t s3sMockTests) DeleteAll(prefix string) error {
	if err := s3t.testsCb(s3t, "DeleteAll", prefix); err != nil {
		return err.(error)
	}
	return nil
}

type s3sMockTests struct {
	parent  IS3Session
	testsCb func(args ...any) interface{}
}

func (s3t s3sMockTests) Check() error { return nil }

func NewS3sMockTests(parent IS3Session, testsCb func(args ...any) interface{}) IS3Session {
	return &s3sMockTests{testsCb: testsCb, parent: parent}
}
