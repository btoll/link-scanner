# LinkScanner

```
Usage of link-scanner:
  -dir string
        Optional.  Searches every file in the directory for a match.  Non-recursive.
  -filename string
        Optional.  Takes precedence over directory searches.
  -filetype .html
        Only searches files of this type.  Include the period, i.e., .html (default ".md")
  -q    Optional.  Turns on quiet mode.
  -v    Optional.  Turns on verbose mode.
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

