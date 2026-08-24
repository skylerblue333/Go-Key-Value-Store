# Security

Sky KV is an engineering-beta, process-local service.

Current controls include bounded request/key/value sizes, strict JSON decoding, bounded key cardinality, HTTP timeouts, concurrency-safe state, dependency vulnerability scanning, and non-root container execution.

The service does not currently provide authentication, authorization, TLS termination, tenant isolation, persistence, encryption at rest, distributed quotas, audit-log durability, or production secret management. Deploy it only behind an appropriate trusted boundary until those controls are implemented and verified.

Report suspected vulnerabilities privately to the repository owner rather than publishing exploit details in a public issue.
