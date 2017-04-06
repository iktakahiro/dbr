# fjord

fjord is a Go lang Struct/DatabaseRecord Mapper package.

## Driver support

* MySQL
* PostgreSQL

## Install

```bash
go get "github.com/iktakahiro/fjord"
```

## Getting Started

```go
conn, _ := fjord.Open("mysql", "...")

sess := conn.NewSession(nil)

var suggestion Suggestion

sess.Select("id", "title").
        From("suggestions").
        Where("id = ?", 1).
        Load(&suggestion)
```

## CRUD

CRUD example using the bellow struct.

```go
type Suggestion struct {
    ID int64
    Title string
    Body fjord.NullString
}
```

### SELECT

```go
var suggestion Suggestion

sess.Select("id", "title").
        From("suggestion").
        Where("id = ?", 1).
        Load(&suggestion)
```

You can implement as a method:

```go
func(s *Suggestion) LoadByID(sess *fjors.session) (count int, err error) {
    count, err = sess.Select("id", "title").
        From("suggestion").
        Where("id = ?", s.ID).
        Load(&s)
           
    return
}

// suggestion := &Suggestion{ID: 1}
// count, err := suggestion.LoadByID(sess)
```

### INSERT

```go
suggestion := &Suggestion{Title: "Gopher", Body: "I love Go."}

sess.InsertInto("suggestion").
        Columns("title", "body").
        Record(suggestion).
        Exec()
```

As a method:

```go
func(s *Suggestion) Save(sess *fjors.session) (err error) {
    err = sess.InsertInto("suggestion").
        Columns("title", "body").
        Record(s).
        Exec()
           
    return
}
```

You can also set a values to insert directly:

```go
sess.InsertInto("suggestion").
        Columns("title", "body").
        Values("Gopher", "I love Go.").
        Exec()
```

Inserting multiple records:

```go
sess.InsertInto("suggestion").
        Columns("title", "body").
        Record(suggestion1).
        Record(suggestion2).
        Exec()
```

### UPDATE

```go
sess.Update("suggestions").
    Set("title", "Gopher").
    Set("body", "We love go.").
    Where("id = ?", 1).
    Exec()
```

`SetMap()` is helpful when you need to update multiple columns.

```go
setMap := map[string]interface{}{
    "title": "Gopher",
    "body": "We love go."
}

sess.Update("suggestion).
    SetMap(setMap).
    Where("id = ?", 1).
    Exec()
```

### DELETE

```go
sess.DeleteFrom("suggestion").
        Where("id = ?", 1).
        Exec()
```

`Soft Delete` does not implemented, use `Update()` manually.

```go
sess.Update("suggestion").
    Set("deleted_at", time.Now().Unix()).
    Where("id = ?", 1).
    Exec()
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

### Load database values to struct fields

```go
// columns are mapped by tag then by field
type Suggestion struct {
    ID int64  // id
    Title string // title
    Url string `db:"-"` // ignored
    secret string // ignored
    Body fjord.NullString `db:"content"` // content
    User User
}

// By default fjord converts CamelCase property names to snake_case column_names
// You can override this with struct tags, just like with JSON tags
type Suggestion struct {
    Id        int64
    Title     fjord.NullString `db:"subject"` // subjects are called titles now
    CreatedAt fjord.NullTime
}

var suggestions []Suggestion
sess.Select("*").From("suggestion").Load(&suggestions)
```

### Table name Alias

```go
sess.Select("s.id", "s.title").
        From(fjord.I("suggestion").As("s")).
        Load(&suggestions)
```

### JOIN

fjord supports many join types:

```go
sess.Select("*").From("suggestion").
  Join("subdomain", "suggestion.subdomain_id = subdomain.id")

sess.Select("*").From("suggestion").
  LeftJoin("subdomain", "suggestions.subdomain_id = subdomain.id")

sess.Select("*").From("suggestion").
  RightJoin("subdomain", "suggestion.subdomain_id = subdomain.id")

sess.Select("*").From("suggestion").
  FullJoin("subdomain", "suggestion.subdomain_id = subdomain.id")
```

You can join on multiple tables:

```go
sess.Select("*").From("suggestion").
  Join("subdomain", "suggestion.subdomain_id = subdomain.id").
  Join("account", "subdomain.account_id = account.id")
```

Combination of JOIN and Aliases

```go
sess.Select("s.id", "s.title", "a.name").
        From(fjord.I("suggestion").As("s")).
        Left(fjord.I("account").As("a"), "s.account_id = a.id")
```

### Quoting/escaping identifiers (e.g. table and column names)

```go
fjord.I("suggestions.id") // `suggestions`.`id`
```

### Sub Query

```go
sess.Select("count(id)").From(
  fjord.Select("*").From("suggestion").As("count"),
)
```

## IN

```go
ids := []int64{1, 2, 3, 4, 5}
builder.Where("id IN ?", ids) // `id` IN ?
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

### Building WHERE condition

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

### Plain SQL

```go
builder := fjord.SelectBySql("SELECT `title`, `body` FROM `suggestion` ORDER BY `id` ASC LIMIT 10")
```

### JSON Friendly Null* types

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

## gocraft/dbr

fjord is [gocraft/dbr](https://github.com/gorcraft/dbr) fork. gocraft/dbr is a really suitable package in many cases.

I'm deeply grateful to the awesome project.

