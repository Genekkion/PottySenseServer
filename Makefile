sql:
	rm -f sqlite.db
	touch sqlite.db
	sqlite3 sqlite.db < schema.sql

example: sql
	sqlite3 sqlite.db < example.sql
