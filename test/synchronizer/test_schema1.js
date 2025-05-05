Schema.database({
  name: "testdb",
  version: "1.0.0",
  collections: {
    users: Schema.collection({
      name: "users",
      docSchema: Schema.doc({
        id: Schema.string().unique(),
        username: Schema.string(),
        role: Schema.string(),
      }),
    }),
    postMetas: Schema.collection({
      name: "postMetas",
      docSchema: Schema.doc({
        id: Schema.string().unique(),
        title: Schema.string(),
        owner: Schema.string(),
      }),
    }),
  },
});
