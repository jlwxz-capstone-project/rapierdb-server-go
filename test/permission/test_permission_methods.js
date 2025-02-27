Permission.create({
  version: "1.0.0",
  rules: {
    documents: {
      canView: (docId, doc, clientId) => {
        // 所有人都可以查看
        return true;
      },
      canCreate: (docId, doc, clientId) => {
        // 只有登录用户可以创建
        return clientId !== "";
      },
      canUpdate: (docId, doc, clientId) => {
        // 只有文档创建者可以更新
        return doc && doc.creator === clientId;
      },
      canDelete: (docId, doc, clientId) => {
        // 只有管理员和创建者可以删除
        return doc && (doc.creator === clientId || clientId === "admin");
      },
    },
  },
});
