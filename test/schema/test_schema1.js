Schema.database({
  name: "testDB",
  version: "1.0.0",
  collections: {
    users: Schema.collection({
      name: "users",
      docSchema: Schema.doc({
        id: Schema.string().unique().index("hash"),
        name: Schema.string().nullable(),
        age: Schema.number().index("range"),
        isActive: Schema.boolean(),
        tags: Schema.list(Schema.string()),
        profile: Schema.object({
          address: Schema.string(),
          phone: Schema.string().nullable(),
        }),
        createdAt: Schema.date(),
        type: Schema.enum(["admin", "user", "guest"]),
        description: Schema.text().index("fulltext"),
        preferences: Schema.record(Schema.string()),
        categories: Schema.tree(Schema.string()),
        sortedItems: Schema.movableList(Schema.string()),
      }),
    }),
  },
});
