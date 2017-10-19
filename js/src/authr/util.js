'use strict';

import AuthrError from './AuthrError';
import { findIndex, keys, isObject, isArray, isString, isPlainObject } from 'lodash';

const has = (obj, attr) => Object.prototype.hasOwnProperty.call(obj, attr);

export const ALLOW = 'allow';
export const DENY = 'deny';

export const SYM_GET_RULES = Symbol('authr.get_rules');
export const SYM_GET_RSRC_TYPE = Symbol('authr.get_resource_type');
export const SYM_GET_RSRC_ATTR = Symbol('authr.get_resource_attribtues');
export const PRIV = Symbol('authr.priv');

/**
 * Determine if a value is able to be evaluated as a resource
 *
 * @param {mixed} thing
 * @return {Boolean}
 */
export function isResource (thing) {
  if (!isObject(thing)) {
    return false;
  }
  if (!has(thing, SYM_GET_RSRC_TYPE) || !has(thing, SYM_GET_RSRC_ATTR)) {
    return false;
  }
  if (typeof thing[SYM_GET_RSRC_TYPE] !== 'function' || typeof thing[SYM_GET_RSRC_ATTR] !== 'function') {
    return false;
  }
  return true;
}

export function assertIsResource (thing) {
  if (!isResource(thing)) {
    throw new AuthrError('"resource" argument does not implement mandatory resource methods');
  }
}

export function isSubject (thing) {
  if (!isObject(thing)) {
    return false;
  }
  if (!has(thing, SYM_GET_RULES)) {
    return false;
  }
  if (typeof thing[SYM_GET_RULES] !== 'function') {
    return false;
  }
  return true;
}

export function assertIsSubject (thing) {
  if (!isSubject(thing)) {
    throw new AuthrError('"subject" argument does not implement mandatory subject methods');
  }
}

function isValidAccess (val) {
  return val === ALLOW || val === DENY;
}

export function assertValidAccess (val) {
  if (!isValidAccess(val)) {
    throw new AuthrError(`Permission constructor expects '${ALLOW}' or '${DENY}' as the only key in an array, got ${val}`);
  }
}

/**
 * Check if a value is sort of empty
 *
 * @param {mixed} val
 * @return {Boolean}
 */
export function empty (val) {
  if (val === null) {
    return true;
  }
  if (val === undefined) {
    return true;
  }
  if (isArray(val) || isString(val)) {
    return !val.length;
  }
  if (isPlainObject(val)) {
    return !keys(val).length;
  }

  return !val;
}

export function inArrayLooseEquality (needle, haystack) {
  return findIndex(haystack, i => i == needle) >= 0; // eslint-disable-line eqeqeq
}
