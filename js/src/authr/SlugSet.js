'use strict';

import { isPlainObject as isObject, isString, isArray, keys } from 'lodash';
import AuthrError from './AuthrError';
import { PRIV } from './util';

const MODE_BLACKLIST = 0;
const MODE_WHITELIST = 1;
const MODE_WILDCARD = 2;

const NOT = '$not';

export default class SlugSet {
  constructor (spec) {
    const self = this[PRIV] = {
      mode: MODE_WHITELIST,
      items: []
    };
    if (spec === '*') {
      self.mode = MODE_WILDCARD;
    } else {
      if (isObject(spec)) {
        let [ modifier ] = keys(spec);
        if (modifier !== NOT) {
          throw new AuthrError('Malformed slug set');
        }
        self.mode = MODE_BLACKLIST;
        spec = spec[NOT];
      }
      if (isString(spec)) {
        spec = [spec];
      }
      if (!isArray(spec)) {
        throw new AuthrError('SlugSet constructor expects a string, array or object for argument 1');
      }
      self.items = spec;
    }
  }

  contains (needle) {
    const self = this[PRIV];
    if (self.mode === MODE_WILDCARD) {
      return true;
    }
    const doesContain = self.items.includes(needle);
    if (self.mode === MODE_BLACKLIST) {
      return !doesContain;
    }

    return doesContain;
  }

  toJSON () {
    const self = this[PRIV];
    if (self.mode === MODE_WILDCARD) {
      return '*';
    }
    var set = self.items;
    if (set.length === 1) {
      [ set ] = set;
    }
    if (self.mode === MODE_BLACKLIST) {
      set = { [NOT]: set };
    }

    return set;
  }
}
