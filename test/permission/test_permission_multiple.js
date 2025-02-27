Permission.create({
  version: "1.0.0",
  rules: {
    users: {
      canView: (docId, doc, clientId) => {
        return true;
      },
      canCreate: (docId, doc, clientId) => {
        return true;
      },
      canUpdate: (docId, doc, clientId) => {
        return doc && doc.owner === clientId;
      },
      canDelete: (docId, doc, clientId) => {
        return clientId === "admin";
      },
    },
    posts: {
      canView: (docId, doc, clientId) => {
        return doc && doc.isPublic === true;
      },
      canCreate: (docId, doc, clientId) => {
        return clientId !== "";
      },
      canUpdate: (docId, doc, clientId) => {
        return doc && doc.author === clientId;
      },
      canDelete: (docId, doc, clientId) => {
        return doc && (doc.author === clientId || clientId === "admin");
      },
    },
    comments: {
      canView: (docId, doc, clientId) => {
        return true;
      },
      canCreate: (docId, doc, clientId) => {
        return clientId !== "";
      },
      canUpdate: (docId, doc, clientId) => {
        return doc && doc.author === clientId;
      },
      canDelete: (docId, doc, clientId) => {
        return doc && (doc.author === clientId || clientId === "admin");
      },
    },
  },
});
