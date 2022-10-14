# Using coral

    $ go install github.com/spf13/cobra/cobra@latest
    $ cd gocode
    $ cobra init cabri
    
    replace github.com/t-beigbeder/otvl_cabri/gocode/cabri/cmd => ./cabri/cmd
    import "github.com/t-beigbeder/otvl_cabri/gocode/cabri/cmd"
    
    $ gofmt -w -r '"github.com/spf13/cobra" -> "github.com/muesli/coral"' .
    $ gofmt -w -r '"github.com/spf13/cobra/doc" -> "github.com/muesli/coral/doc"' .
    $ gofmt -w -r 'cobra -> coral' .
    $ go mod tidy
    
    $ cd gocode/cabri
    $ cobra add serve
    $ cobra add config
    $ cobra add create -p 'configCmd'
    $ gofmt -w -r '"github.com/spf13/cobra" -> "github.com/muesli/coral"' .
    $ gofmt -w -r '"github.com/spf13/cobra/doc" -> "github.com/muesli/coral/doc"' .
    $ gofmt -w -r 'cobra -> coral' .
