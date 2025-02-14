const indexTypes = {
  none: "none",
  hash: "hash",
  range: "range",
  fulltext: "fulltext",
};

var ErrInvalidNullable = new Error("Invalid nullable");
var ErrInvalidUnique = new Error("Invalid unique");
var ErrInvalidIndexType = new Error("Invalid index type");
var ErrInvalidEnumValues = new Error("Invalid enum values");
var ErrInvalidShape = new Error("Invalid shape");
var ErrInvalidListItemSchema = new Error("Invalid list item schema");
var ErrInvalidRecordValueSchema = new Error("Invalid record value schema");
var ErrInvalidTreeNodeSchema = new Error("Invalid tree node schema");
var ErrInvalidDocFields = new Error("Invalid doc fields");
var ErrInvalidCollectionParams = new Error("Invalid collection params");
var ErrInvalidDatabaseParams = new Error("Invalid database params");

var anySymbol = { type: "any" };
var booleanSymbol = { type: "boolean" };
var dateSymbol = { type: "date" };
var enumSymbol = { type: "enum" };
var listSymbol = { type: "list" };
var movableListSymbol = { type: "movableList" };
var numberSymbol = { type: "number" };
var objectSymbol = { type: "object" };
var recordSymbol = { type: "record" };
var stringSymbol = { type: "string" };
var textSymbol = { type: "text" };
var treeSymbol = { type: "tree" };
var docSymbol = { type: "doc" };
var collectionSymbol = { type: "collection" };
var databaseSymbol = { type: "database" };

var valueAndContainerSymbols = [
  anySymbol,
  booleanSymbol,
  dateSymbol,
  enumSymbol,
  listSymbol,
  movableListSymbol,
  numberSymbol,
  objectSymbol,
  recordSymbol,
  stringSymbol,
  textSymbol,
  treeSymbol,
  // docSymbol, collectionSymbol, databaseSymbol are not allowed
];

var Schema = {};

Schema.any = function () {
  var _nullable = false;
  var ret = {
    _symbol: anySymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "any",
        nullable: _nullable,
      };
    },
  };
  return ret;
};

Schema.boolean = function () {
  var _nullable = false;
  var _unique = false;
  var _indexType = "none";
  var ret = {
    _symbol: booleanSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    unique: function (unique) {
      if (unique === undefined) {
        _unique = true;
      } else if (typeof unique === "boolean") {
        _unique = unique;
      } else {
        throw ErrInvalidUnique;
      }
      return ret;
    },
    index: function (indexType) {
      if (indexType in indexTypes) {
        _indexType = indexType;
      } else {
        throw ErrInvalidIndexType;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "boolean",
        nullable: _nullable,
        unique: _unique,
        indexType: _indexType,
      };
    },
  };
  return ret;
};

Schema.date = function () {
  var _nullable = false;
  var _unique = false;
  var _indexType = "none";
  var ret = {
    _symbol: dateSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    unique: function (unique) {
      if (unique === undefined) {
        _unique = true;
      } else if (typeof unique === "boolean") {
        _unique = unique;
      } else {
        throw ErrInvalidUnique;
      }
      return ret;
    },
    index: function (indexType) {
      if (indexType in indexTypes) {
        _indexType = indexType;
      } else {
        throw ErrInvalidIndexType;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "date",
        nullable: _nullable,
        unique: _unique,
        indexType: _indexType,
      };
    },
  };
  return ret;
};

Schema.enum = function (values) {
  if (!Array.isArray(values)) {
    throw ErrInvalidEnumValues;
  }

  for (var i = 0; i < values.length; i++) {
    if (typeof values[i] !== "string") {
      throw ErrInvalidEnumValues;
    }
  }

  var _nullable = false;
  var _unique = false;
  var _indexType = "none";
  var ret = {
    _symbol: enumSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    unique: function (unique) {
      if (unique === undefined) {
        _unique = true;
      } else if (typeof unique === "boolean") {
        _unique = unique;
      } else {
        throw ErrInvalidUnique;
      }
      return ret;
    },
    index: function (indexType) {
      if (indexType in indexTypes) {
        _indexType = indexType;
      } else {
        throw ErrInvalidIndexType;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "enum",
        values: values,
        nullable: _nullable,
        unique: _unique,
        indexType: _indexType,
      };
    },
  };
  return ret;
};

Schema.list = function (itemSchemaBuilder) {
  if (
    !("_symbol" in itemSchemaBuilder) ||
    !valueAndContainerSymbols.includes(itemSchemaBuilder._symbol)
  ) {
    throw ErrInvalidListItemSchema;
  }

  var _nullable = false;
  var ret = {
    _symbol: listSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "list",
        itemSchema: itemSchemaBuilder.toJSON(),
        nullable: _nullable,
      };
    },
  };
  return ret;
};

Schema.movableList = function (itemSchemaBuilder) {
  if (
    !("_symbol" in itemSchemaBuilder) ||
    !valueAndContainerSymbols.includes(itemSchemaBuilder._symbol)
  ) {
    throw ErrInvalidListItemSchema;
  }

  var _nullable = false;
  var ret = {
    _symbol: movableListSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "movableList",
        itemSchema: itemSchemaBuilder.toJSON(),
        nullable: _nullable,
      };
    },
  };
  return ret;
};

Schema.number = function () {
  var _nullable = false;
  var _unique = false;
  var _indexType = "none";
  var ret = {
    _symbol: numberSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    unique: function (unique) {
      if (unique === undefined) {
        _unique = true;
      } else if (typeof unique === "boolean") {
        _unique = unique;
      } else {
        throw ErrInvalidUnique;
      }
      return ret;
    },
    index: function (indexType) {
      if (indexType in indexTypes) {
        _indexType = indexType;
      } else {
        throw ErrInvalidIndexType;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "number",
        nullable: _nullable,
        unique: _unique,
        indexType: _indexType,
      };
    },
  };
  return ret;
};

Schema.object = function (shape) {
  if (typeof shape !== "object") {
    throw ErrInvalidShape;
  }

  for (var key in shape) {
    var val = shape[key];
    if (
      !("_symbol" in val) ||
      !valueAndContainerSymbols.includes(val._symbol)
    ) {
      throw ErrInvalidShape;
    }
  }

  var _nullable = false;
  var ret = {
    _symbol: objectSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    toJSON: function () {
      var shapeJSON = {};
      for (var key in shape) {
        shapeJSON[key] = shape[key].toJSON();
      }

      return {
        type: "object",
        shape: shapeJSON,
        nullable: _nullable,
      };
    },
  };
  return ret;
};

Schema.record = function (valueSchemaBuilder) {
  if (
    !("_symbol" in valueSchemaBuilder) ||
    !valueAndContainerSymbols.includes(valueSchemaBuilder._symbol)
  ) {
    throw ErrInvalidRecordValueSchema;
  }

  var _nullable = false;
  var ret = {
    _symbol: recordSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "record",
        valueSchema: valueSchemaBuilder.toJSON(),
        nullable: _nullable,
      };
    },
  };
  return ret;
};

Schema.string = function () {
  var _nullable = false;
  var _unique = false;
  var _indexType = "none";
  var ret = {
    _symbol: stringSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    unique: function (unique) {
      if (unique === undefined) {
        _unique = true;
      } else if (typeof unique === "boolean") {
        _unique = unique;
      } else {
        throw ErrInvalidUnique;
      }
      return ret;
    },
    index: function (indexType) {
      if (indexType in indexTypes) {
        _indexType = indexType;
      } else {
        throw ErrInvalidIndexType;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "string",
        nullable: _nullable,
        unique: _unique,
        indexType: _indexType,
      };
    },
  };
  return ret;
};

Schema.text = function () {
  var _nullable = false;
  var _indexType = "none";
  var ret = {
    _symbol: textSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    index: function (indexType) {
      // hash and range are not allowed for text
      if (["none", "fulltext"].includes(indexType)) {
        _indexType = indexType;
      } else {
        throw ErrInvalidIndexType;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "text",
        nullable: _nullable,
        indexType: _indexType,
      };
    },
  };
  return ret;
};

Schema.tree = function (treeNodeSchemaBuilder) {
  if (
    !("_symbol" in treeNodeSchemaBuilder) ||
    !valueAndContainerSymbols.includes(treeNodeSchemaBuilder._symbol)
  ) {
    throw ErrInvalidTreeNodeSchema;
  }

  var _nullable = false;
  var ret = {
    _symbol: treeSymbol,
    nullable: function (nullable) {
      if (nullable === undefined) {
        _nullable = true;
      } else if (typeof nullable === "boolean") {
        _nullable = nullable;
      } else {
        throw ErrInvalidNullable;
      }
      return ret;
    },
    toJSON: function () {
      return {
        type: "tree",
        treeNodeSchema: treeNodeSchemaBuilder.toJSON(),
        nullable: _nullable,
      };
    },
  };
  return ret;
};

Schema.doc = function (fields) {
  if (typeof fields !== "object") {
    throw ErrInvalidDocFields;
  }

  for (var key in fields) {
    var val = fields[key];
    if (!valueAndContainerSymbols.includes(val._symbol)) {
      throw ErrInvalidDocFields;
    }
  }

  var ret = {
    _symbol: docSymbol,
    toJSON: function () {
      var fieldsJSON = {};
      for (var key in fields) {
        fieldsJSON[key] = fields[key].toJSON();
      }

      return {
        type: "doc",
        fields: fieldsJSON,
      };
    },
  };
  return ret;
};

Schema.collection = function (params) {
  if (
    typeof params !== "object" ||
    !("name" in params) ||
    !("docSchema" in params) ||
    Object.keys(params).length !== 2
  ) {
    throw ErrInvalidCollectionParams;
  }

  var _name = params.name;
  var _docSchema = params.docSchema;

  if (typeof _name !== "string") {
    throw ErrInvalidCollectionParams;
  }

  if (
    typeof _docSchema !== "object" ||
    !("_symbol" in _docSchema) ||
    _docSchema._symbol !== docSymbol
  ) {
    throw ErrInvalidCollectionParams;
  }

  var ret = {
    _symbol: collectionSymbol,
    name: _name,
    toJSON: function () {
      return {
        type: "collection",
        name: _name,
        docSchema: _docSchema.toJSON(),
      };
    },
  };
  return ret;
};

Schema.database = function (params) {
  if (
    typeof params !== "object" ||
    !("name" in params) ||
    !("version" in params) ||
    !("collections" in params) ||
    Object.keys(params).length !== 3
  ) {
    throw ErrInvalidDatabaseParams;
  }

  var _name = params.name;
  var _version = params.version;
  var _collections = params.collections;

  if (typeof _name !== "string") {
    throw ErrInvalidDatabaseParams;
  }

  if (typeof _version !== "string") {
    throw ErrInvalidDatabaseParams;
  }

  if (typeof _collections !== "object") {
    throw ErrInvalidDatabaseParams;
  }

  for (var key in _collections) {
    var val = _collections[key];
    if (
      typeof val !== "object" ||
      !("_symbol" in val) ||
      val._symbol !== collectionSymbol ||
      val.name !== key
    ) {
      throw ErrInvalidDatabaseParams;
    }
  }

  var ret = {
    _symbol: databaseSymbol,
    toJSON: function () {
      var collectionsJSON = {};
      for (var key in _collections) {
        collectionsJSON[key] = _collections[key].toJSON();
      }

      return {
        type: "database",
        name: _name,
        version: _version,
        collections: collectionsJSON,
      };
    },
  };
  return ret;
};
