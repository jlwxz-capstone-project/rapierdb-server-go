Permission.create({
  version: "1.0.0",
  rules: {
    postMetas: {
      /**
       * 用于判断 ID 为 clientId 的客户端能否查看一个文档
       * @param {Object} params - 参数对象
       * @param {string} params.docId 这个文档的 id
       * @param {*} params.doc 这个文档
       * @param {string} params.clientId 当前客户端的 id
       * @param {*} params.db 数据库的只读代理，允许在权限中查询文档。应该尽可能少用，
       *               因为权限校验函数会被频繁地调用，查询文档尤其是复杂的查询是较慢的操作。
       */
      canView: ({ docId, doc, clientId, db }) => {
        // 查询当前用户
        const client = db.users.findOne({
          filter: eq(field("data/id"), clientId),
        });
        // 如果当前用户不存在，则没有权限
        if (!client) return false;
        // 只允许文档的 owner 或者角色为 admin 的客户端查看
        return doc.data.owner === clientId || client.data.role === "admin";
      },
      /**
       * 用于判断 ID 为 clientId 的客户端能否创建一个文档
       * @param {Object} params - 参数对象
       * @param {string} params.docId 这个文档的 id
       * @param {*} params.newDoc 要插入的新文档
       * @param {string} params.clientId 当前客户端的 id
       * @param {*} params.db 数据库的代理，允许在权限中创建文档。应该尽可能少用，
       *               因为权限校验函数会被频繁地调用，创建文档尤其是复杂的创建是较慢的操作。
       */
      canCreate: ({ docId, newDoc, clientId, db }) => {
        return true; // 所有客户端都可以创建文档
      },
      /**
       * 用于判断 ID 为 clientId 的客户端能否更新一个文档
       * @param {Object} params - 参数对象
       * @param {string} params.docId 这个文档的 id
       * @param {*} params.newDoc 新的文档数据
       * @param {*} params.oldDoc 旧的文档数据
       * @param {string} params.clientId 当前客户端的 id
       * @param {*} params.db 数据库的代理，允许在权限中更新文档。应该尽可能少用，
       *               因为权限校验函数会被频繁地调用，更新文档尤其是复杂的更新是较慢的操作。
       */
      canUpdate: ({ docId, newDoc, oldDoc, clientId, db }) => {
        // 检查是否修改了文档所有者
        if (newDoc.data.owner !== oldDoc.data.owner) return false;
        // 查询当前用户
        const client = db.users.findOne({
          filter: eq(field("data/id"), clientId),
        });
        // 如果当前用户不存在，则没有权限
        if (!client) return false;
        return oldDoc.data.owner === clientId || client.data.role === "admin";
      },
      /**
       * 用于判断 ID 为 clientId 的客户端能否删除一个文档
       * @param {Object} params - 参数对象
       * @param {string} params.docId 这个文档的 id
       * @param {*} params.doc 这个文档
       * @param {string} params.clientId 当前客户端的 id
       * @param {*} params.db 数据库的代理，允许在权限中删除文档。应该尽可能少用，
       *               因为权限校验函数会被频繁地调用，删除文档尤其是复杂的删除是较慢的操作。
       */
      canDelete: ({ docId, doc, clientId, db }) => {
        // 只允许文档的 owner 或者角色为 admin 的客户端删除
        const client = db.users.findOne((user) => user.id === clientId);
        return doc.data.owner === clientId || client.role === "admin";
      },
    },
  },
});
