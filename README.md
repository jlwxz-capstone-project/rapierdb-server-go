# RapierDB

RapierDB provides you an **offline-capable**, **reactive** and **realtime** client database.

## Before & After

Changes for Client-Side Developers:

| Before⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀                                          | After⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀                                            | What's happening behind the scenes                           |
| ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| Manually initiate GET requests to fetch data from the server | Directly query the client-side database                      | The database automatically registers this query with the server, synchronizes data, and updates the result set when synchronization completes (ensuring results are reactive & real-time). |
| Manually initiate POST requests to modify server-side data   | Directly modify the client-side database                     | If there's no network connection, the modification is committed to the client-side database but not yet submitted to the server. Once connected to the server, the database automatically submits pending transactions. If successful, the backend service returns an ACK to the client that submitted the transaction and pushes the updated documents to other clients based on their subscribed queries, ensuring real-time updates. |
| Manually maintain server state cache on the client (using tools like React Query) | No need to maintain cache, just use the client-side database directly | Specifically, clients can choose whether to enable optimistic updates. If enabled, the client-side database is updated first to minimize UI latency, then the transaction is submitted to the server. If the server rejects the transaction, a rollback operation is executed. Otherwise, the transaction is submitted directly to the server, and the client-side database is updated after server confirmation. |
| To ensure state is real-time, manually receive server events via SSE or Websocket, then invalidate corresponding caches based on received events | Handled entirely by the framework, developers don't need to worry about it | Queries to the client-side database are reactive, and as long as connected to the backend service, the client-side database is synchronized with the backend database in real-time, ensuring that query results are always up-to-date. |
| Manually store states that need persistence in localStorage / IndexedDB, then restore state when the application loads | The client-side database is persistent by default            | -                                                            |

Changes for Server-Side Developers:

| Before⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀                                         | After⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀                                          | What's happening behind the scenes                           |
| ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| Manually parse WAL (such as MySQL's binlog) to identify affected queries, notify clients via SSE or Websocket, then re-execute queries upon receiving new GET requests from clients | Real-time related work is completely handled by the framework | The backend service maintains all queries registered by each client and calculates which document sets each client is "interested in." For each newly committed transaction, the Event-Reduce algorithm calculates which client queries will be affected, then pushes new documents to the affected clients. |
| Implement permission verification at the API layer           | Use our authentication module to verify user identity, then use JavaScript to write permission verification rules implementing permission control at the data layer | For transaction submission requests from clients, the request first passes through the authentication module to verify identity. If successful, the transaction and identity information are passed through the authorization function. If that succeeds, the transaction is allowed to be submitted. |
| Encapsulate database functionality as APIs for frontend use  | Frontend directly operates on the client-side database, whose data automatically synchronizes with the server-side database - no need to write backend APIs anymore! | -                                                            |
