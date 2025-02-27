Permission.create({
  version: "1.0.0",
  rules: {
    users: {
      canView: (docId, doc, clientId) => {
        // 只有当文档的 owner 是当前客户端或者角色为 admin 时才能查看
        return doc && (doc.owner === clientId || clientId === "admin");
      },
      canCreate: (docId, doc, clientId) => {
        // 所有客户端都可以创建文档
        return true;
      },
      canUpdate: (docId, doc, clientId) => {
        // 只有文档的 owner 或者角色为 admin 的客户端才能更新
        return doc && (doc.owner === clientId || clientId === "admin");
      },
      canDelete: (docId, doc, clientId) => {
        // 只有角色为 admin 的客户端才能删除
        return clientId === "admin";
      },
    },
  },
});
