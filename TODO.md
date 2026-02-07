This is TODO list for AI agent:
- ensure each bug fixes or features doesn't create new bugs
- ensure each bug fixes or features is tested with real database
- ensure each bug fixes or features have explanations

The human side:
- i installed docker and docker-compose for you, the AI agent, to use it to install mongodb, redis, or whatever needed


TODO list:
[ ] Error panic: ':wedding_id' in new path '/api/v1/weddings/:wedding_id/guests' conflicts with existing wildcard ':id' in existing prefix '/api/v1/weddings/:id'
[ ] Failed to ensure indexes        {"error": "failed to create system_analytics _id index: (InvalidIndexSpecificationOption) The field 'unique' is not valid for an _id index specification. Specification: { key: { _id: 1 }, name: \"_id_1\", unique: true, v: 2 }"}
[ ] Failed to ensure indexes        {"error": "failed to create wedding_analytics _id index: (InvalidIndexSpecificationOption) The field 'unique' is not valid for an _id index specification. Specification: { key: { _id: 1 }, name: \"_id_1\", unique: true, v: 2 }"}