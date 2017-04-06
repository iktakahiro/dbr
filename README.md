# fjord (database records)

fjord provides additions to Go's database/sql for super fast performance and convenience.

## Getting Started

```go
// create a connection (e.g. "postgres", "mysql", or "sqlite3")
conn, _ := fjord.Open("postgres", "...")

// create a session for each business unit of execution (e.g. a web request or goworkers job)
sess := conn.NewSession(nil)

// get a record
var suggestion Suggestion
sess.Select("id", "title").From("suggestions").Where("id = ?", 1).Load(&suggestion)

// JSON-ready, with fjord.Null* types serialized like you want
json.Marshal(&suggestion)
```

## Feature highlights

### Use a Sweet Query Builder or use Plain SQL

fjord supports both.

Sweet Query Builder:
```go
stmt := fjord.Select("title", "body").
    From("suggestions").
    OrderBy("id").
    Limit(10)
```

Plain SQL:

```go
builder := fjord.SelectBySql("SELECT `title`, `body` FROM `suggestions` ORDER BY `id` ASC LIMIT 10")
```

### Amazing instrumentation with session

All queries in fjord are made in the context of a session. This is because when instrumenting your app, it's important to understand which business action the query took place in. See gocraft/health for more detail.

Writing instrumented code is a first-class concern for gocraft/fjord. We instrument each query to emit to a gocraft/health-compatible EventReceiver interface.

### Faster performance than using database/sql directly
Every time you call database/sql's db.Query("SELECT ...") method, under the hood, the mysql driver will create a prepared statement, execute it, and then throw it away. This has a big performance cost.

fjord doesn't use prepared statements. We ported mysql's query escape functionality directly into our package, which means we interpolate all of those question marks with their arguments before they get to MySQL. The result of this is that it's way faster, and just as secure.

Check out these [benchmarks](https://github.com/tyler-smith/golang-sql-benchmark).

### IN queries that aren't horrible
Traditionally, database/sql uses prepared statements, which means each argument in an IN clause needs its own question mark. fjord, on the other hand, handles interpolation itself so that you can easily use a single question mark paired with a dynamically sized slice.

```go
ids := []int64{1, 2, 3, 4, 5}
builder.Where("id IN ?", ids) // `id` IN ?
```

### JSON Friendly
Every try to JSON-encode a sql.NullString? You get:
```json
{
    "str1": {
        "Valid": true,
        "String": "Hi!"
    },
    "str2": {
        "Valid": false,
        "String": ""
  }
}
```

Not quite what you want. fjord has fjord.NullString (and the rest of the Null* types) that encode correctly, giving you:

```json
{
    "str1": "Hi!",
    "str2": null
}
```

### Inserting multiple records

```go
sess.InsertInto("suggestions").Columns("title", "body").
  Record(suggestion1).
  Record(suggestion2)
```

### Updating records

```go
sess.Update("suggestions").
    Set("title", "Gopher").
    Set("body", "I love go.").
    Where("id = ?", 1)
```

### Transactions

```go
tx, err := sess.Begin()
if err != nil {
  return err
}
defer tx.RollbackUnlessCommitted()

// do stuff...

return tx.Commit()
```

### Load database values to variables

Querying is the heart of fjord.

* Load(&any): load everything!

```go
// columns are mapped by tag then by field
type Suggestion struct {
    ID int64  // id, will be autoloaded by last insert id
    Title string // title
    Url string `db:"-"` // ignored
    secret string // ignored
    Body fjord.NullString `db:"content"` // content
    User User
}

// By default fjord converts CamelCase property names to snake_case column_names
// You can override this with struct tags, just like with JSON tags
// This is especially helpful while migrating from legacy systems
type Suggestion struct {
    Id        int64
    Title     fjord.NullString `db:"subject"` // subjects are called titles now
    CreatedAt fjord.NullTime
}

var suggestions []Suggestion
sess.Select("*").From("suggestions").Load(&suggestions)
```

### Join multiple tables

fjord supports many join types:

```go
sess.Select("*").From("suggestions").
  Join("subdomains", "suggestions.subdomain_id = subdomains.id")

sess.Select("*").From("suggestions").
  LeftJoin("subdomains", "suggestions.subdomain_id = subdomains.id")

sess.Select("*").From("suggestions").
  RightJoin("subdomains", "suggestions.subdomain_id = subdomains.id")

sess.Select("*").From("suggestions").
  FullJoin("subdomains", "suggestions.subdomain_id = subdomains.id")
```

You can join on multiple tables:

```go
sess.Select("*").From("suggestions").
  Join("subdomains", "suggestions.subdomain_id = subdomains.id").
  Join("accounts", "subdomains.accounts_id = accounts.id")
```

### Quoting/escaping identifiers (e.g. table and column names)

```go
fjord.I("suggestions.id") // `suggestions`.`id`
```

### Subquery

```go
sess.Select("count(id)").From(
  fjord.Select("*").From("suggestions").As("count"),
)
```

### Union

```go
fjord.Union(
  fjord.Select("*"),
  fjord.Select("*"),
)

fjord.UnionAll(
  fjord.Select("*"),
  fjord.Select("*"),
)
```

Union can be used in subquery.

### Alias/AS

* SelectStmt

```go
fjord.Select("*").From("suggestions").As("count")
```

* Identity

```go
fjord.I("suggestions").As("s")
```

* Union

```go
fjord.Union(
  fjord.Select("*"),
  fjord.Select("*"),
).As("u1")

fjord.UnionAll(
  fjord.Select("*"),
  fjord.Select("*"),
).As("u2")
```

### Building arbitrary condition

One common reason to use this is to prevent string concatenation in a loop.

* And
* Or
* Eq
* Neq
* Gt
* Gte
* Lt
* Lte

```go
fjord.And(
  fjord.Or(
    fjord.Gt("created_at", "2015-09-10"),
    fjord.Lte("created_at", "2015-09-11"),
  ),
  fjord.Eq("title", "hello world"),
)
```

### Built with extensibility

The core of fjord is interpolation, which can expand `?` with arbitrary SQL. If you need a feature that is not currently supported,
you can build it on your own (or use `fjord.Expr`).

To do that, the value that you wish to be expaned with `?` needs to implement `fjord.Builder`.

```go
type Builder interface {
    Build(Dialect, Buffer) error
}
```

## Driver support

* MySQL
* PostgreSQL

## gocraft/dbr

fjord is [gocraft/dbr](https://github.com/gorcraft/dbr) fork. gocraft/dbr is a really suitable package in many cases.

I'm deeply grateful to the awesome project.
