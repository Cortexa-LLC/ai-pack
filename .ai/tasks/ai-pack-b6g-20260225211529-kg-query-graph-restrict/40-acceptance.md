# Query Graph Tool Restrictions

The query_graph tool has been updated to restrict operations to read-only queries only. The following mutations are now rejected:
- CREATE
- MERGE
- DELETE
- SET

## Allowed Queries

Only the following commands are permitted:
- MATCH
- RETURN

Ensure any attempts to execute mutation commands will result in an error, indicating that write operations are not permitted. 

This change is aimed at improving the safety and integrity of data operations within the system, preventing potential accidental modifications from user queries.

## Testing

Thorough testing was conducted to validate that only read operations are executed successfully, while write operations are denied. Users are encouraged to abide by these new restrictions to ensure the tool remains functional and secure.