'use strict';

import test from 'ava';
import ConditionSet from '../../src/authr/ConditionSet';
import { GET_RESOURCE_TYPE, GET_RESOURCE_ATTRIBUTE } from '../../src/authr';

test('unknown logical conjunctions throws error', t => {
  t.throws(() => {
    new ConditionSet({ // eslint-disable-line no-new
      $xor: [['@id', '=', '1'], ['@type', '=', 'root']]
    });
  });
});

test('weird construction values throws error', t => {
  t.throws(() => {
    new ConditionSet(8); // eslint-disable-line no-new
  });
  t.throws(() => {
    new ConditionSet({ $and: { $or: ['what', 'are', 'you', 'doing?!'] } }); // eslint-disable-line no-new
  });
});

test('normal construction gives a normal ConditionSet', t => {
  var attrs = {};
  var rsrc = {
    [GET_RESOURCE_TYPE]: () => 'user',
    [GET_RESOURCE_ATTRIBUTE]: k => {
      return attrs[k] || null;
    }
  };

  var cs = new ConditionSet([
    ['@type', '~=', 'root'],
    {
      $or: [
        ['@id', '=', '1'],
        ['@id', '=', '888']
      ]
    }
  ]);

  attrs['id'] = '44';
  t.false(cs.evaluate(rsrc));

  attrs['type'] = 'ROOT';
  t.false(cs.evaluate(rsrc));

  attrs['id'] = '1';
  t.true(cs.evaluate(rsrc));

  attrs['id'] = '888';
  t.true(cs.evaluate(rsrc));

  attrs['id'] = '90';
  t.false(cs.evaluate(rsrc));
});

test('ConditionSet will skip over random falsy values', t => {
  var rsrc = {
    [GET_RESOURCE_TYPE]: () => 'user',
    [GET_RESOURCE_ATTRIBUTE]: k => {
      var attrs = { id: '5' };
      return attrs[k] || null;
    }
  };
  var cs = new ConditionSet([
    [],
    ['@id', '=', '5'],
    false
  ]);
  t.true(cs.evaluate(rsrc));
});

test('OR evaluations can short-circuit if needed', t => {
  var rsrc = {
    [GET_RESOURCE_TYPE]: () => 'user',
    [GET_RESOURCE_ATTRIBUTE]: k => {
      var attrs = {
        one: 'two',
        three: 'four'
      };
      if (k === 'three') {
        t.fail('short circuit did not work, ConditionSet continued evaluating');
      }
      return attrs[k] || null;
    }
  };

  var cs = new ConditionSet({
    $or: [
      ['@one', '=', 'two'],
      ['@three', '!=', 'four']
    ]
  });

  t.true(cs.evaluate(rsrc));
});

test('AND evaluations can short-circuit if needed', t => {
  var rsrc = {
    [GET_RESOURCE_TYPE]: () => 'user',
    [GET_RESOURCE_ATTRIBUTE]: k => {
      var attrs = {
        one: 'two',
        three: 'four'
      };
      if (k === 'three') {
        t.fail('short circuit did not work, ConditionSet continued evaluating');
      }
      return attrs[k] || null;
    }
  };

  var cs = new ConditionSet([
    ['@one', '=', 'five'],
    ['@three', '=', 'four']
  ]);

  t.false(cs.evaluate(rsrc));
});

test('vacuous truth', t => {
  var rsrc = {
    [GET_RESOURCE_TYPE]: () => 'zone',
    [GET_RESOURCE_ATTRIBUTE]: k => null
  };

  t.true(new ConditionSet([]).evaluate(rsrc));
});

test('evaluating non-resource throws error', t => {
  var rsrc = {
    id: '5',
    type: 'resource, trust me, im a dolphin'
  };

  t.throws(() => {
    new ConditionSet(['@id', '=', '5']).evaluate(rsrc);
  });
});

test('ConditionSet toString', t => {
  var cs = new ConditionSet({
    $and: [
      ['@id', '=', '5'],
      ['@type', '=', 'admin']
    ]
  });
  t.is(JSON.stringify(cs), '[["@id","=","5"],["@type","=","admin"]]');

  cs = new ConditionSet({
    $or: [
      ['@type', '=', 'root'],
      ['@id', '=', '1']
    ]
  });
  t.is(JSON.stringify(cs), '{"$or":[["@type","=","root"],["@id","=","1"]]}');
});
