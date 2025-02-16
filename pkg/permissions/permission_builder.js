var Permission = {};

const ErrInvalidPermissionParams = new Error("invalid permission params");

Permission.create = function (permission) {
  var _version = permission.version;
  var _rules = permission.rules;

  if (!_version || !_rules) {
    throw ErrInvalidPermissionParams;
  }
};
