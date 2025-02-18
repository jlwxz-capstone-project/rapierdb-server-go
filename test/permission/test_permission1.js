Permission.create({
  version: "1.0.0",
  rules: {
    users: {
      canView: (doc, ctx) => {
        return true;
      },
      canCreate: (doc, ctx) => {
        return true;
      },
      canUpdate: (doc, ctx) => {
        return true;
      },
      canDelete: (doc, ctx) => {
        return true;
      },
    },
    postMetas: {
      canView: (doc, ctx) => {
        return true;
      },
      canCreate: (doc, ctx) => {
        return true;
      },
      canUpdate: (doc, ctx) => {
        return true;
      },
      canDelete: (doc, ctx) => {
        return true;
      },
    },
    postContents: {
      canView: (doc, ctx) => {
        return true;
      },
      canCreate: (doc, ctx) => {
        return true;
      },
      canUpdate: (doc, ctx) => {
        return true;
      },
      canDelete: (doc, ctx) => {
        return true;
      },
    },
  },
});
