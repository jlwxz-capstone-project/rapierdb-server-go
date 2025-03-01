Permission.create({
  version: "1.0.0",
  rules: {
    users: {
      /**
       * 用于判断 ID 为 clientId 的客户端能否查看一个文档
       * @param {*} docId 这个文档的 id
       * @param {*} doc 这个文档
       * @param {*} clientId 当前客户端的 id
       * @param {*} db 数据库的只读代理，允许在权限中查询文档。应该尽可能少用，
       *               因为权限校验函数会被频繁地调用，查询文档尤其是复杂的查询是较慢的操作。
       */
      canView: (docId, doc, clientId, db) => {
        return doc.data.owner === clientId;
      },
      /**
       * 用于判断 ID 为 clientId 的客户端能否创建一个文档
       * @param {*} docId 这个文档的 id
       * @param {*} doc 这个文档
       * @param {*} clientId 当前客户端的 id
       * @param {*} db 数据库的代理，允许在权限中创建文档。应该尽可能少用，
       *               因为权限校验函数会被频繁地调用，创建文档尤其是复杂的创建是较慢的操作。
       */
      canCreate: (docId, doc, clientId, db) => {
        // 所有客户端都可以创建文档
        return true;
      },
      /**
       * 用于判断 ID 为 clientId 的客户端能否更新一个文档
       * @param {*} docId 这个文档的 id
       * @param {*} doc 这个文档
       * @param {*} clientId 当前客户端的 id
       * @param {*} db 数据库的代理，允许在权限中更新文档。应该尽可能少用，
       *               因为权限校验函数会被频繁地调用，更新文档尤其是复杂的更新是较慢的操作。
       */
      canUpdate: (docId, doc, clientId, db) => {
        // 只有文档的 owner 或者角色为 admin 的客户端才能更新
      },
      /**
       * 用于判断 ID 为 clientId 的客户端能否删除一个文档
       * @param {*} docId 这个文档的 id
       * @param {*} doc 这个文档
       * @param {*} clientId 当前客户端的 id
       * @param {*} db 数据库的代理，允许在权限中删除文档。应该尽可能少用，
       *               因为权限校验函数会被频繁地调用，删除文档尤其是复杂的删除是较慢的操作。
       */
      canDelete: (docId, doc, clientId) => {
        // 只有角色为 admin 的客户端才能删除
        return clientId === "admin";
      },
    },
  },
});
