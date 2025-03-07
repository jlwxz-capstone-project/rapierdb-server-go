Schema.database({
  name: "testDB",
  version: "1.0.0",
  collections: {
    users: Schema.collection({
      name: "users",
      docSchema: Schema.doc({
        id: Schema.string().unique().index("hash"),
        name: Schema.maxLength(255).minLength(1).string().nullable(),
        age: Schema.number().min(0).max(100).index("range"),
        isActive: Schema.boolean(),
        tags: Schema.list(Schema.string()),
        profile: Schema.object({
          address: Schema.string(),
          phone: Schema.string().nullable(),
        }),
        createdAt: Schema.date().index("range"),
        type: Schema.enum(["admin", "user", "guest"]),
        description: Schema.text().index("fulltext"),
        preferences: Schema.record(Schema.string()),
        categories: Schema.tree(Schema.string()),
        sortedItems: Schema.movableList(Schema.string()),
      }),
    }),
  },
});
