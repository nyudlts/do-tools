## do-tools

go-rstar-cli

This repo contains command-line-interface (CLI) code for interacting
with elements of an ArchivesSpace instance. 


#### Running
```
$ do-tools -h     # prints out available commands

# example command 
$ do-tools do refresh --ao-uri /repositories/11/archival_objects/237 -c /path/to/cfg.yml -e test 
```

#### Testing
The current test suite assumes that you have a file named `config.yaml` 
in the `cmd/testsupport` directory with an environment named `test`.  
There is a template file in the `cmd/testsupport` directory.  
Note that `cmd/testsupport/config.yaml` is ignored by Git via the  
`.gitignore` file to avoid accidentally committing application secrets.  

Therefore, to run tests, you can simply `cd` into `cmd/` directory and run  
`$ go test -v`

