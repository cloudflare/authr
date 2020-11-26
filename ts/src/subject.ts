import { isObject } from "lodash";
import Rule from "./rule";

export const SYM_GET_RULES = Symbol("authr.subject_get_rules");

interface ISubject {
  [SYM_GET_RULES](): Rule[];
}

export function isSubject(v?: any): v is ISubject {
  if (!isObject(v)) {
    return false;
  }
  return v.hasOwnProperty(SYM_GET_RULES);
}

export default ISubject;
