import isString from 'lodash.isstring';
import AuthrError from './AuthrError';
import Rule from './Rule';
import {
  assertIsResource,
  assertIsSubject,
  ALLOW,
  DENY,
  SYM_GET_RULES,
  SYM_GET_RSRC_TYPE,
  SYM_GET_RSRC_ATTR
} from './util';

function can (subject, action, resource) {
  assertIsSubject(subject);
  assertIsResource(resource);
  if (!isString(action)) {
    throw new AuthrError('"action" must be a string');
  }
  const rt = resource[SYM_GET_RSRC_TYPE]();
  const rules = subject[SYM_GET_RULES]();
  for (let rule of rules) {
    if (!(rule instanceof Rule)) {
      throw new AuthrError('List of subject permissions contained a value that is not a Permission');
    }
    if (!rule.resourceTypes().contains(rt)) {
      continue;
    }
    if (!rule.actions().contains(action)) {
      continue;
    }
    if (!rule.conditions().evaluate(resource)) {
      continue;
    }

    let type = rule.access();
    if (type === ALLOW) {
      return true;
    } else if (type === DENY) {
      return false;
    }

    // unknown type!
    throw new Error(`Permission type set to unknown value: '${type}'`);
  }

  // default to "deny all"
  return false;
}

export {
  can,
  Rule,
  AuthrError,
  SYM_GET_RULES as GET_RULES,
  SYM_GET_RSRC_TYPE as GET_RESOURCE_TYPE,
  SYM_GET_RSRC_ATTR as GET_RESOURCE_ATTRIBUTE
};
