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

## Git `pre-commit` Hook

Download to `.git/hooks` directory and make executable:

```
wget -P .git/hooks/ \
    https://raw.githubusercontent.com/btoll/dotfiles/master/git-hub/hooks/pre-commit
chmod 755 .git/hooks/pre-commit
```

Create the new directory and add the hook:

```
mkdir .git/hooks/pre-commit.d
wget -P .git/hooks/pre-commit.d \
    https://raw.githubusercontent.com/btoll/dotfiles/master/git-hub/hooks/pre-commit.d/link-scanner.sh
```

Finally, add the hook as a local config:

```
git config --local --add hooks.pre-commit.hook link-scanner.sh
```

## TODO

- Add verbose flag.
    + Will show what files are ignored.
- Add quiet flag.
    + Will suppress the `FAILED` line with number of failures.
    + Will enable easier `stdout` parsing by subsequent programs in a pipeline.
- Defaults to only showing failures.  This is determined to be `401`s.  Let user decide.
- Add color?

