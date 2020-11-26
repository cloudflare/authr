import { isString } from "lodash";
import AuthrError from "./authrError";
import IResource, { SYM_GET_RSRC_ATTR, SYM_GET_RSRC_TYPE } from "./resource";
import Rule, { Access } from "./rule";
import ISubject, { SYM_GET_RULES } from "./subject";
import { runtimeAssertIsResource, runtimeAssertIsSubject } from "./util";

function can(subject: ISubject, action: string, resource: IResource): boolean {
  runtimeAssertIsSubject(subject);
  runtimeAssertIsResource(resource);
  if (!isString(action)) {
    throw new AuthrError('"action" must be a string');
  }
  const rules = subject[SYM_GET_RULES]();
  const rt = resource[SYM_GET_RSRC_TYPE]();
  for (let rule of rules) {
    if (!rule.resourceTypes().contains(rt)) {
      continue;
    }
    if (!rule.actions().contains(action)) {
      continue;
    }
    if (!rule.conditions().evaluate(resource)) {
      continue;
    }
    const access = rule.access();
    switch (access) {
      case Access.ALLOW:
        return true;
      case Access.DENY:
        return false;
    }
    throw new Error(`Rule access set to unknown value: '${access}'`);
  }
  // default to "deny all"
  return false;
}

export {
  ISubject,
  IResource,
  can,
  Rule,
  AuthrError,
  SYM_GET_RULES as GET_RULES,
  SYM_GET_RSRC_TYPE as GET_RESOURCE_TYPE,
  SYM_GET_RSRC_ATTR as GET_RESOURCE_ATTRIBUTE,
};
