# Reproducer for rows affected in dqlite

## Usage

Run `make build` and then run `dqlite-bug-reproducer`

Dqlite output is the following:

```
$ dqlite-bug-reproducer
Starting...
Using dqlite
Ready
Running statements
ERROR: update expected 1 row affected, got 0
```

Sqlite3 output is the following:

```
$ dqlite-bug-reproducer sqlite
Starting...
Using sqlite3
Running statements
Success
```
