package cabridss

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/internal"
	"github.com/t-beigbeder/otvl_cabri/gocode/packages/ufpath"
	"io"
	"os"
	"strings"
	"time"
)

const MIN_TIME int64 = -9223372036854775808
const MAX_TIME int64 = 9223372036854775807
const MAX_META_SIZE = 100000

type WriteCloserCb func(err error, size int64, ch string)

// Dss is the Data Storage System interface.
//
// Any  Dss should implement this interface.
type Dss interface {
	// Mkns creates a namespace in the Dss, return an error if any happens
	//
	// npath is the full namespace without leading or trailing slash
	// mtime is the last modification POSIX time
	// children are the children names, a trailing slash denotes a namespace, else regular content
	// acl is the access control List to the namespace
	Mkns(npath string, mtime int64, children []string, acl []ACLEntry) error

	// Updatens updates a namespace in the Dss, return an error if any happens
	//
	// npath is the full namespace without leading or trailing slash
	// mtime is the last modification POSIX time
	// children are the children names, a trailing slash denotes a namespace, else regular content
	// acl is the access control List to the namespace
	Updatens(npath string, mtime int64, children []string, acl []ACLEntry) error

	// Lsns lists a namespace's content, return it or an error if any happens
	//
	// npath is the full namespace without leading or trailing slash
	//
	// returns:
	// - children names, a trailing slash denotes a namespace, else regular content
	// - err error if any happens
	Lsns(npath string) (children []string, err error)

	// IsDuplicate checks if content's checksum ch already exists in DSS
	//
	// returns duplicate status and an error if any happens
	IsDuplicate(ch string) (bool, error)

	// GetContentWriter creates content for writing
	//
	// npath is the full namespace + name without leading slash
	// mtime is the last modification POSIX time
	// acl is the access control List to the content
	// cb if not nil is a callback called when writer is closed
	//
	// returns:
	// - a writer to provide the content
	// - err error if any happens
	GetContentWriter(npath string, mtime int64, acl []ACLEntry, cb WriteCloserCb) (io.WriteCloser, error)

	// GetContentReader opens content for reading
	//
	// npath is the full namespace + name without leading slash
	//
	// returns:
	// - a reader to retrieve the content
	// - err error if any happens
	GetContentReader(npath string) (io.ReadCloser, error)

	// Remove removes a namespace (and recursively its children) or some content
	//
	// npath is the full namespace + name without leading slash, trailing slash indicates it is a namespace
	//
	// returns:
	// - err error if any happens
	Remove(npath string) error

	// GetMeta gets a namespace or some content metadata
	//
	// npath is the full namespace + name without leading slash, trailing slash indicates it is a namespace
	//
	// returns:
	// - the metadata
	// - err error if any happens
	GetMeta(npath string, getCh bool) (IMeta, error)

	// SetCurrentTime injects arbitrary current time for tests
	// does nothing on fsy DSS
	SetCurrentTime(time int64)

	// SetMetaMockCbs injects arbitrary json [un]marshal for tests if build tag constraint enables
	SetMetaMockCbs(cbs *MetaMockCbs)

	// SetAfs injects abstract FS for tests if build tag constraint enables
	SetAfs(tfs afero.Fs)

	// GetAfs retrieves it
	GetAfs() afero.Fs

	// Close can be necessary to perform final cleanup or index synchronization
	Close() error

	// SetSu enables superuser access for synchro
	SetSu()

	// SuEnableWrite enables physical write access for fsy: DSS in case SetSu is active
	//
	// npath is the full namespace + name without leading slash, trailing slash indicates it is a namespace
	//
	// returns:
	// - err error if any happens
	SuEnableWrite(npath string) error
}

type UnixUTC int64

func (t UnixUTC) String() string {
	if int64(t) != MIN_TIME && int64(t) != MAX_TIME {
		sec, nano := internal.Nano2SecNano(int64(t))
		return time.Unix(sec, nano).UTC().Format("2006-01-02T15:04:05")
	} else {
		return "....-..-..T..:..:.."
	}
}

type UnixNanoUTC int64

func (t UnixNanoUTC) String() string {
	if int64(t) != MIN_TIME && int64(t) != MAX_TIME {
		sec, nano := internal.Nano2SecNano(int64(t))
		return time.Unix(sec, nano).UTC().Format("2006-01-02T15:04:05.99999999")
	} else {
		return "....-..-..T..:..:.."
	}
}

// TimeResolution values are "s" seconds, "m" minutes, "h" hours, "d" days
type TimeResolution string

func (tr TimeResolution) NanoSeconds() int64 {
	if tr == "s" {
		return 1e9
	} else if tr == "m" {
		return 60 * 1e9
	} else if tr == "h" {
		return 3600 * 1e9
	} else if tr == "d" {
		return 24 * 3600 * 1e9
	}
	panic(fmt.Sprintf("TimeResolution %s is inconsistent", tr))
}

func (tr TimeResolution) Align(ns int64) int64 {
	d := tr.NanoSeconds()
	r := (ns / d) * d
	if ns < 0 {
		r -= d
	}
	return r
}

type HistoryInfo struct {
	Start int64 // start time POSIX of the history entry
	End   int64 // end time POSIX of the history entry
	HMeta Meta  // the  metadata
}

func (hi HistoryInfo) String() string {
	return fmt.Sprintf("%s/%s %12d %s %s", UnixUTC(hi.Start), UnixUTC(hi.End), hi.HMeta.Size, hi.HMeta.Ch, hi.HMeta.Path)
}

type AuditIndexInfo struct {
	Error string // "IndexInternal", "IndexMissing", "StorageMissing", "Inconsistent"
	Err   error  // origin error
	Time  int64  // the time of the entry in the DSS
	Bytes []byte // the metadata
}

func (aii AuditIndexInfo) String() string {
	fe := func(e string) string { return e }
	return fmt.Sprintf("%s %12d %s (%s: %v)", UnixUTC(aii.Time), len(aii.Bytes), internal.BytesToSha256Str(aii.Bytes), fe(aii.Error), aii.Err)
}

type SIHnIt struct {
	Hn string `json:"hn"`
	It int64  `json:"it"`
}

type StorageInfo struct {
	Path2Meta     map[string][]byte           `json:"path2Meta"`
	Path2HnIt     map[string]SIHnIt           `json:"path2HnIt"` // Local meta data
	ExistingCs    map[string]bool             `json:"existingCs"`
	ExistingEcs   map[string]bool             `json:"existingEcs"`
	Path2Content  map[string]string           `json:"path2Content"`
	Path2CContent map[string]string           `json:"path2CContent"`
	Path2Error    map[string]error            `json:"path2Error"`
	XLMetas       map[string]map[int64][]byte `json:"xlmetas"` // Local meta data
	XRMetas       map[string]map[int64][]byte `json:"xrmetas"` // Remote meta data
}

type HistoryChunk struct {
	Start int64 `json:"start"` // period start time POSIX aligned to resolution
	End   int64 `json:"end"`   // period end time POSIX aligned to resolution
	Count int   `json:"count"` // number of history updates in the time period
}

func (hc HistoryChunk) String() string {
	return fmt.Sprintf("%s/%s %8d", UnixUTC(hc.Start), UnixUTC(hc.End), hc.Count)
}

// HDss is the Data Storage System interface for DSS with history support.
type HDss interface {
	Dss

	// GetHistory gets the history of the npath entry as a map of entry state sorted by time
	//
	// npath is the full namespace + name without leading slash, trailing slash indicates it is a namespace
	// recursive requests the service to recursively get the history of all namespace children
	// resolution "s" for seconds, "m" for minutes, "h" for hours, "d" for days summarizes the history with given resolution
	//
	// returns:
	// - the history (inclusive times when the entry is visible) for all entries
	// - err error if any happens
	GetHistory(npath string, recursive bool, resolution string) (map[string][]HistoryInfo, error)

	// RemoveHistory removes history entries for a given time period
	//
	// it must be noted that removing a parent history may cause children to be removed for a larger period of time
	//
	// npath is the full namespace + name without leading slash, trailing slash indicates it is a namespace
	// recursive requests the service to recursively remove the history of all namespace children,
	// evaluate don't remove, just report work to be done
	// start is the inclusive index time above which entries must be removed, zero meanning all past entries
	// end is the inclusive index time below which entries must be removed, zero meaning all future entries
	//
	// returns:
	// - the history (inclusive times when the entry is removed) for all entries
	// - err error if any happens
	RemoveHistory(npath string, recursive, evaluate bool, start, end int64) (map[string][]HistoryInfo, error)

	// GetIndex provides the DSS index or nil
	GetIndex() Index

	// DumpIndex for debug
	DumpIndex() string

	// GetRepoId provides the DSS repoId or ""
	GetRepoId() string

	// IsEncrypted tells if repository is encrypted
	IsEncrypted() bool

	// IsRepoEncrypted tells if repository configuration is set to encrypted
	IsRepoEncrypted() bool

	// AuditIndex compares the DSS index with meta and content actually stored
	AuditIndex() (map[string][]AuditIndexInfo, error)

	// ScanStorage scans the DSS storage and loads meta and content sha256 sum
	//
	// checksum checks content checksums
	// purge removes unreferenced content from the repository
	// purgeHidden removes hidden meta and content from the repository
	ScanStorage(checksum, purge, purgeHidden bool) (StorageInfo, *ErrorCollector)

	// GetHistoryChunks returns history chunks loaded from local index
	//
	// resolution s, m, h, d from seconds to days
	//
	// returns:
	// - the DSS activity periods sorted by time
	// - err error if any happens
	GetHistoryChunks(resolution string) ([]HistoryChunk, error)

	// Reindex scans the DSS storage and loads meta and content sha256 sum into the index
	Reindex() (StorageInfo, *ErrorCollector)
}

var appFs = afero.NewOsFs()

func checkDir(root string) error {
	fi, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("not a directory: %s", root)
	}
	return nil
}

func checkNpath(npath string) error {
	if len(npath) > 0 && (npath[0] == '/' || npath[len(npath)-1] == '/') {
		return fmt.Errorf("namespace %s should not have leading or trailing slash", npath)
	}
	return nil
}

func checkNCpath(npath string) (isNS bool, ipath string, err error) {
	isNS = npath == "" || (len(npath) > 0 && npath[len(npath)-1] == '/')
	if isNS && npath != "" {
		ipath = npath[:len(npath)-1]
	} else {
		ipath = npath
	}
	if len(ipath) > 0 && ipath[0] == '/' {
		err = fmt.Errorf("namespace %s should not have leading slash", npath)
	}
	return
}

func checkName(name string) error {
	sn := name
	if len(sn) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	if name[len(sn)-1] == '/' {
		sn = name[0 : len(sn)-1]
	}
	if strings.ContainsAny(sn, "/\n") {
		return fmt.Errorf("name %s should not contain LF or other slash than trailing one", name)
	}
	return nil
}

func checkNames(names []string) error {
	var errNames []string
	uniqueNames := make(map[string]bool)
	for _, name := range names {
		if err := checkName(name); err != nil {
			errNames = append(errNames, name)
		}
		if _, ok := uniqueNames[name]; ok {
			errNames = append(errNames, name)
		}
		uniqueNames[name] = true
	}
	if errNames != nil {
		return fmt.Errorf("name(s) %v should not be empty, contain LF or other slash than trailing one, or appear more than once", errNames)
	}
	return nil
}

func checkMknsArgs(npath string, children []string, acl []ACLEntry) error {
	if err := checkNpath(npath); err != nil {
		return err
	}
	if err := checkNames(children); err != nil {
		return err
	}
	return nil
}

func checkMkcontentArgs(npath string, acl []ACLEntry) error {
	if err := checkNpath(npath); err != nil {
		return err
	}
	return nil
}

type CreateNewParams struct {
	ConfigPassword string                                                      // if not "" master password used to encrypt client configuration
	ConfigDir      string                                                      // if not "" path to the user's configuration directory
	Create         bool                                                        // perform CreateXxx or NewXxx?
	DssType        string                                                      // fsy, olf, obs, smf
	Root           string                                                      // fsy, olf, smf
	Size           string                                                      // if olf: s,m,l
	LocalPath      string                                                      // fsy, obs, smf (Root assumed if olf or smf)
	Encrypted      bool                                                        // all but fsy: enable repository encryption
	GetIndex       func(config DssBaseConfig, localPath string) (Index, error) // see DssBaseConfig
	Lsttime        int64                                                       // all but fsy: if not zero is the upper time of entries retrieved in it
	Aclusers       []string                                                    // all but fsy: if not nil is a List of ACL users for access check
	Endpoint       string                                                      // obs: AWS S3 or Openstack Swift endpoint, eg "https://s3.gra.cloud.ovh.net"
	Region         string                                                      // obs: AWS S3  or Openstack Swift region, eg "GRA"
	AccessKey      string                                                      // obs: AWS S3 access key (Openstack Swift must generate it)
	SecretKey      string                                                      // obs: AWS S3 secret key (Openstack Swift must generate it)
	Container      string                                                      // obs: AWS S3 bucket or Openstack Swift container
	RedLimit       int                                                         // all: reducer limit or 0
}

func CreateOrNewDss(params CreateNewParams) (dss Dss, err error) {
	if params.DssType == "olf" {
		localPath := params.LocalPath
		if localPath == "" {
			localPath = params.Root
		}
		config := OlfConfig{
			DssBaseConfig: DssBaseConfig{ConfigDir: params.ConfigDir, ConfigPassword: params.ConfigPassword, LocalPath: localPath, GetIndex: params.GetIndex, Encrypted: params.Encrypted, ReducerLimit: params.RedLimit},
			Root:          params.Root, Size: params.Size,
		}
		if params.Create {
			dss, err = CreateOlfDss(config)
		} else {
			dss, err = NewOlfDss(config, params.Lsttime, params.Aclusers)
		}
		return dss, err
	}
	if params.DssType == "obs" {
		config := ObsConfig{
			DssBaseConfig: DssBaseConfig{ConfigDir: params.ConfigDir, ConfigPassword: params.ConfigPassword, LocalPath: params.LocalPath, GetIndex: params.GetIndex, Encrypted: params.Encrypted, ReducerLimit: params.RedLimit},
			Endpoint:      params.Endpoint,
			Region:        params.Region,
			AccessKey:     params.AccessKey,
			SecretKey:     params.SecretKey,
			Container:     params.Container,
		}
		if params.Create {
			dss, err = CreateObsDss(config)
		} else {
			dss, err = NewObsDss(config, params.Lsttime, params.Aclusers)
		}
		return dss, err
	}
	if params.DssType == "smf" {
		localPath := params.LocalPath
		if localPath == "" {
			localPath = params.Root
		}
		config := ObsConfig{
			DssBaseConfig: DssBaseConfig{ConfigDir: params.ConfigDir, ConfigPassword: params.ConfigPassword, LocalPath: localPath, GetIndex: params.GetIndex, Encrypted: params.Encrypted, ReducerLimit: params.RedLimit},
			Endpoint:      params.Endpoint,
			Region:        params.Region,
			AccessKey:     params.AccessKey,
			SecretKey:     params.SecretKey,
			Container:     params.Container,
			GetS3Session: func() IS3Session {
				return NewS3sMockFs(localPath, nil)
			},
		}
		if params.Create {
			dss, err = CreateObsDss(config)
		} else {
			dss, err = NewObsDss(config, params.Lsttime, params.Aclusers)
		}
		return dss, err
	}
	return nil, fmt.Errorf("in CreateOrNewDss: DSS type %s not (yet) supported", params.DssType)
}

// GetPIndex provides the buntdb index with the localPath
func GetPIndex(bc DssBaseConfig, localPath string) (Index, error) {
	return NewPIndex(ufpath.Join(bc.LocalPath, "index.bdb"), bc.Unlock, bc.AutoRepair)
}
