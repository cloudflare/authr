'use strict';

import AuthrError from './AuthrError';
import Condition from './Condition';
import {
  empty,
  assertIsResource,
  PRIV
} from './util';
import isArray from 'lodash.isarray';
import isPlainObject from 'lodash.isplainobject';
import keys from 'lodash.keys';
import isBoolean from 'lodash.isboolean';

const LOGICAL_AND = '$and';
const LOGICAL_OR = '$or';
const IMPLIED_CONJUNCTION = LOGICAL_AND;

export default class ConditionSet {
  constructor (spec = null) {
    const self = this[PRIV] = {
      conjunction: LOGICAL_AND,
      evaluators: []
    };
    if (isPlainObject(spec)) {
      let [ conjunction ] = keys(spec);
      if (conjunction !== LOGICAL_AND && conjunction !== LOGICAL_OR) {
        throw new AuthrError(`Unknown condition set conjunction: ${conjunction}`);
      }
      self.conjunction = conjunction;
      spec = spec[self.conjunction];
    }
    if (!isArray(spec)) {
      throw new AuthrError('ConditionSet only takes an object or array during construction');
    }
    for (let rawEvaluator of spec) {
      if (empty(rawEvaluator)) {
        continue;
      }
      if (isArray(rawEvaluator) && rawEvaluator.length === 3 && typeof rawEvaluator[0] === 'string') {
        self.evaluators.push(new Condition(...rawEvaluator));
      } else {
        self.evaluators.push(new ConditionSet(rawEvaluator));
      }
    }
  }

  evaluate (resource) {
    assertIsResource(resource);
    const self = this[PRIV];
    var result = true; // Vacuous truth: https://en.wikipedia.org/wiki/Vacuous_truth
    for (let evaluator of self.evaluators) {
      let evalResult = evaluator.evaluate(resource);
      if (!isBoolean(evalResult)) {
        throw new AuthrError('Evaluator returned a non-boolean value');
      }
      if (self.conjunction === LOGICAL_OR) {
        if (evalResult) {
          return true; // short circuit
        }
        result = false;
      }
      if (self.conjunction === LOGICAL_AND) {
        if (!evalResult) {
          return false; // short circuit
        }
        result = true;
      }
    }
    return result;
  }

  toJSON () {
    const self = this[PRIV];
    var out = self.evaluators.map(e => e.toJSON());
    if (self.conjunction !== IMPLIED_CONJUNCTION) {
      out = { [self.conjunction]: out };
    }
    return out;
  }
}
