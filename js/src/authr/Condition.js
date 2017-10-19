'use strict';

import has from 'lodash.has';
import isArray from 'lodash.isarray';
import isString from 'lodash.isstring';
import {
  assertIsResource,
  inArrayLooseEquality,
  SYM_GET_RSRC_ATTR,
  PRIV
} from './util';
import AuthrError from './AuthrError';

const OPERATORS = {
  '=': (left, right) => left == right, // eslint-disable-line eqeqeq
  '!=': (left, right) => left != right, // eslint-disable-line eqeqeq
  '~=': (left, right) => {
    let pleft = '^';
    let pright = '$';
    right = `${right}`;
    if (right.startsWith('*')) {
      pleft = '';
    }
    if (right.endsWith('*')) {
      pright = '';
    }
    return RegExp(`${pleft}${right.replace('*', '')}${pright}`, 'i').test(left);
  },
  '~': (left, right) => RegExp(right).test(left),
  '~*': (left, right) => RegExp(right, 'i').test(left),
  '!~': (left, right) => !RegExp(right).test(left),
  '!~*': (left, right) => !RegExp(right, 'i').test(left),
  '$in': (left, right) => {
    if (!isArray(right)) {
      return false;
    }
    return inArrayLooseEquality(left, right);
  },
  '$nin': (left, right) => {
    if (!isArray(right)) {
      return false;
    }
    return !inArrayLooseEquality(left, right);
  }
};

/**
 * Determine if the value passed is referring to an attribute on the resource
 * or is just the literal value.
 *
 * @param {object} resource
 * @param {string} value
 * @return {mixed}
 */
function determineValue (resource, value) {
  if (isString(value) && value.length > 1) {
    if (value.charAt(0) === '@') {
      return resource[SYM_GET_RSRC_ATTR](value.substr(1));
    }
    if (value.substr(0, 2) === '\\@') {
      return value.substr(1);
    }
  }
  return value;
}

export default class Condition {
  /**
   * @param {any} left
   * @param {string} operator
   * @param {any} right
   * @return Condition
   */
  constructor (left, operator, right) {
    if (!has(OPERATORS, operator)) {
      throw new AuthrError('Unknown condition operator');
    }
    this[PRIV] = { operator, left, right };
  }

  evaluate (resource) {
    assertIsResource(resource);
    const self = this[PRIV];
    return OPERATORS[self.operator](
      determineValue(resource, self.left),
      determineValue(resource, self.right)
    );
  }

  toJSON () {
    const { left, operator, right } = this[PRIV];
    return [left, operator, right];
  }
}
