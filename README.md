# LinkScanner

```
Usage of ./link-scanner:
  -dir string
        Optional.  Searches every file in the directory for a match.  Non-recursive. (default ".")
  -filename string
        Optional.  Takes precedence over directory searches.
  -filetype string
        Only searches files of this type. (default ".md")
```

## TODO

- Add verbose flag.
    + Will show what files are ignored.
- Add quiet flag.
    + Will suppress the `FAILED` line with number of failures.
    + Will enable easier `stdout` parsing by subsequent programs in a pipeline.
- Defaults to only showing failures.  This is determined to be `401`s.  Let user decide.
- Add color?

