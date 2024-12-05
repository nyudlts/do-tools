package cmd

// $0 ao get-root --ao-uri|-a          // finds the root archival object for the ao-uri argument, returning itself if there are no ancestors
// $0 do create   --ao-uri|-a <ao URI> --file-version|-f <file URI> --use-statement|-u <use statement>
// $0 do update   --ao-uri|-a <ao URI> --old-file-version|-o <file URI to replace> --new-file-version|-n <new file URI value> --use-statement|-u <new FV use statement>
// $0 do refresh  --ao-uri|-a <ao URI> // updates the metadata of all DOs attached to the AO
