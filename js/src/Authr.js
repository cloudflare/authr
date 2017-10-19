import { isString } from 'lodash';
import AuthrError from './authr/AuthrError';
import Rule from './authr/Rule';
import {
  assertIsResource,
  assertIsSubject,
  ALLOW,
  DENY,
  SYM_GET_RULES,
  SYM_GET_RSRC_TYPE,
  SYM_GET_RSRC_ATTR
} from './authr/util';

export const SUBJECT_PERMISSIONS = Symbol.for('authr.get_rules');
export const RESOURCE_TYPE = Symbol.for('authr.get_resource_type');
export const RESOURCE_ATTR = Symbol.for('authr.get_resource_attribute');

export default class Authr {
  static can (subject, action, resource) {
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
}

export {
  Rule,
  AuthrError,
  SYM_GET_RULES as GET_RULES,
  SYM_GET_RSRC_TYPE as GET_RESOURCE_TYPE,
  SYM_GET_RSRC_ATTR as GET_RESOURCE_ATTRIBUTE
};
