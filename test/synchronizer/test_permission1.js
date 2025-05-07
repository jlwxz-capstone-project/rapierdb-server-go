Permission.create({
  version: "1.0.0",
  rules: {
    users: {
      canView: ({ docId, doc, clientId, db }) => {
        if (clientId === "admin") return true; // admin can view all users
        return doc.id === clientId; // user himself can view his own info
      },
      canCreate: ({ docId, newDoc, clientId, db }) => {
        // only admin can create new user
        return clientId === "admin";
      },
      canUpdate: ({ docId, newDoc, oldDoc, clientId, db }) => {
        if (clientId === "admin") return true; // admin can update all users
        return false; // don't allow users to update info
      },
      canDelete: ({ docId, doc, clientId, db }) => {
        if (clientId === "admin") return true;
        return doc.id === clientId; // user himself can delete his own info
      },
    },
    postMetas: {
      canView: ({ docId, doc, clientId, db }) => {
        const query = {
          filter: eq(field("id"), clientId),
        };
        const client = db.users.findOne(query);
        if (!client) return false;
        return doc.owner === clientId || client.role === "admin";
      },
      canCreate: ({ docId, newDoc, clientId, db }) => {
        return true;
      },
      canUpdate: ({ docId, newDoc, oldDoc, clientId, db }) => {
        if (newDoc.owner !== oldDoc.owner) return false;
        const client = db.users.findOne({
          filter: eq(field("id"), clientId),
        });
        if (!client) return false;
        return oldDoc.owner === clientId || client.role === "admin";
      },
      canDelete: ({ docId, doc, clientId, db }) => {
        const client = db.users.findOne((user) => user.id === clientId);
        return doc.owner === clientId || client.role === "admin";
      },
    },
  },
});
