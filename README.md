DBGormen

This is small Golang MYSQL client with codegeneration. 

dbgormen/generator - files for codegenerator. 

Usage:

go run dbgorme.go -name=/path/to/your/file.go

DBGormen looks for structures with comment // dbe and generates methods for this structures:

- createTable
- Create
- Query
- Update
- Delete

Every generated method gets connection to DB as argument.

With //dbe comment you can set tablename: {"table": "name"}

In tags for structure fields can be set:

- primary_key
- not_null
- change column name for field
- use - for preventing column generation and using this field in connection to DB.