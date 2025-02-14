schema = Schema.database({
  name: "test",
  version: "1.0.0",
  collections: {
    users: Schema.collection({
      name: "users",
      docSchema: Schema.doc({
        userId: Schema.string().unique(),
        username: Schema.string(),
        email: Schema.string(),
        passwordHash: Schema.string(),
      }),
    }),
    postMetas: Schema.collection({
      postId: Schema.string().unique(),
      title: Schema.string(),
      caption: Schema.string(),
      authorId: Schema.string(),
    }),
    postContents: Schema.collection({
      postId: Schema.string(),
      content: Schema.string(),
    }),
  },
});
