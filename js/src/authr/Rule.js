import SlugSet from './SlugSet';
import ConditionSet from './ConditionSet';
import AuthrError from './AuthrError';
import {
  assertValidAccess,
  ALLOW,
  DENY,
  PRIV
} from './util';
import { isPlainObject } from 'lodash';

const RSRC_TYPE = 'rsrc_type';
const RSRC_MATCH = 'rsrc_match';
const ACTION = 'action';

export default class Rule {
  constructor (access, spec, meta = null) {
    const self = this[PRIV] = {
      access: ALLOW,
      segments: {
        [RSRC_TYPE]: null,
        [RSRC_MATCH]: null,
        [ACTION]: null
      },
      meta: null
    };
    if (spec === 'all') {
      spec = {
        [RSRC_TYPE]: '*',
        [RSRC_MATCH]: [],
        [ACTION]: '*'
      };
    }
    if (!isPlainObject(spec)) {
      throw new AuthrError('"spec" must be a plain object');
    }
    assertValidAccess(access);
    self.access = access;
    if (meta) {
      self.meta = meta;
    }
    for (let seg in spec) {
      let segspec = spec[seg];
      switch (seg) {
        case RSRC_TYPE:
        case ACTION:
          self.segments[seg] = new SlugSet(segspec);
          break;
        case RSRC_MATCH:
          self.segments[seg] = new ConditionSet(segspec);
          break;
      }
    }
  }

  static create (spec) {
    if (!isPlainObject(spec)) {
      throw new AuthrError('"spec" must be a plain object');
    }
    return new Rule(spec.access, spec.where, spec.$meta);
  }

  static allow (spec, meta = null) {
    return new Rule(ALLOW, spec, meta);
  }

  static deny (spec, meta = null) {
    return new Rule(DENY, spec, meta);
  }

  access () {
    return this[PRIV].access;
  }

  resourceTypes () {
    const self = this[PRIV];
    if (!self.segments[RSRC_TYPE]) {
      throw new AuthrError('Cannot retrieve resource type segment');
    }

    return self.segments[RSRC_TYPE];
  }

  conditions () {
    const self = this[PRIV];
    if (!self.segments[RSRC_MATCH]) {
      throw new AuthrError('Cannot retrieve resource match segment');
    }

    return self.segments[RSRC_MATCH];
  }

  actions () {
    const self = this[PRIV];
    if (!self.segments[ACTION]) {
      throw new AuthrError('Cannot retrieve actions segment');
    }

    return self.segments[ACTION];
  }

  toJSON () {
    const self = this[PRIV];
    const raw = {
      access: self.access,
      where: {
        [ACTION]: this.actions().toJSON(),
        [RSRC_TYPE]: this.resourceTypes().toJSON(),
        [RSRC_MATCH]: this.conditions().toJSON()
      }
    };
    if (self.meta) {
      raw.$meta = self.meta;
    }
    return raw;
  }

  toString () {
    return JSON.stringify(this);
  }
}
