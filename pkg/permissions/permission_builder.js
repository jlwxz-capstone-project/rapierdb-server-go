var Permission = {};

const ErrInvalidPermissionParams = new Error("invalid permission params");

const requiredRuleKeys = ["canView", "canCreate", "canUpdate", "canDelete"];

Permission.create = function (permission) {
  var _version = permission.version;
  var _rules = permission.rules;

  if (
    !_version ||
    typeof _version !== "string" ||
    !_rules ||
    typeof _rules !== "object"
  ) {
    throw ErrInvalidPermissionParams;
  }

  for (const key in _rules) {
    const rule = _rules[key];
    if (!rule || typeof rule !== "object") {
      throw ErrInvalidPermissionParams;
    }

    for (const key in rule) {
      if (!requiredRuleKeys.includes(key)) {
        throw ErrInvalidPermissionParams;
      }

      const value = rule[key];
      if (typeof value !== "function") {
        throw ErrInvalidPermissionParams;
      }
    }
  }

  return {
    version: _version,
    rules: _rules,
  };
};
