'use strict';

import test from 'ava';
import ConditionSet from '../../src/Permission/ConditionSet';

test('unknown logical conjunctions throws error', t => {
  t.throws(() => {
    ConditionSet.create({
      $xor: [['@id', '=', '1'], ['@type', '=', 'root']]
    });
  });
});

test('weird construction values throws error', t => {
  t.throws(() => {
    ConditionSet.create(8);
  });
  t.throws(() => {
    ConditionSet.create({ $and: { $or: ['what', 'are', 'you', 'doing?!'] } });
  });
});

test('normal construction gives a normal ConditionSet', t => {
  var attrs = {};
  var rsrc = {
    [Symbol.for('permission.resource_type')]: 'user',
    [Symbol.for('permission.resource_attr')]: k => {
      return attrs[k] || null;
    }
  };

  var cs = ConditionSet.create([
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
    [Symbol.for('permission.resource_type')]: 'user',
    [Symbol.for('permission.resource_attr')]: k => {
      var attrs = { id: '5' };
      return attrs[k] || null;
    }
  };
  var cs = ConditionSet.create([
    [],
    ['@id', '=', '5'],
    false
  ]);
  t.true(cs.evaluate(rsrc));
});

test('OR evaluations can short-circuit if needed', t => {
  var rsrc = {
    [Symbol.for('permission.resource_type')]: 'user',
    [Symbol.for('permission.resource_attr')]: k => {
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

  var cs = ConditionSet.create({
    $or: [
      ['@one', '=', 'two'],
      ['@three', '!=', 'four']
    ]
  });

  t.true(cs.evaluate(rsrc));
});

test('AND evaluations can short-circuit if needed', t => {
  var rsrc = {
    [Symbol.for('permission.resource_type')]: 'user',
    [Symbol.for('permission.resource_attr')]: k => {
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

  var cs = ConditionSet.create([
    ['@one', '=', 'five'],
    ['@three', '=', 'four']
  ]);

  t.false(cs.evaluate(rsrc));
});

test('vacuous truth', t => {
  var rsrc = {
    [Symbol.for('permission.resource_type')]: 'zone',
    [Symbol.for('permission.resource_attr')]: k => null
  };

  t.true(ConditionSet.create([]).evaluate(rsrc));
});

test('evaluating non-resource throws error', t => {
  var rsrc = {
    id: '5',
    type: 'resource, trust me, im a dolphin'
  };

  t.throws(() => {
    ConditionSet.create(['@id', '=', '5']).evaluate(rsrc);
  });
});

test('ConditionSet toString', t => {
  var cs = ConditionSet.create({
    $and: [
      ['@id', '=', '5'],
      ['@type', '=', 'admin']
    ]
  });
  t.is(JSON.stringify(cs), '[["@id","=","5"],["@type","=","admin"]]');

  cs = ConditionSet.create({
    $or: [
      ['@type', '=', 'root'],
      ['@id', '=', '1']
    ]
  });
  t.is(JSON.stringify(cs), '{"$or":[["@type","=","root"],["@id","=","1"]]}');
});
