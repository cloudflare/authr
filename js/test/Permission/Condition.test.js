'use strict';

import test from 'ava';
import Condition from '../../src/Permission/Condition';

test('undefined operator throws error', t => {
  t.throws(() => {
    Condition.create('id', '@>', 'one');
  });
});

test('evaluating non-resource throws error', t => {
  t.throws(() => {
    Condition.create('id', '=', '5').evaluate({
      id: '5'
    });
  });
});

test('condition evaluate', t => {
  var resource = {
    [Symbol.for('permission.resource_type')]: 'zone',
    [Symbol.for('permission.resource_attr')]: key => {
      var attrs = {
        id: 123,
        type: 'full',
        kind: 'Pretty',
        user_id: '867',
        lol: '@wut',
        wut: '???'
      };
      return attrs[key] || null;
    }
  };

  t.true(Condition.create('@id', '=', '123').evaluate(resource));
  t.true(Condition.create('@id', '=', 123).evaluate(resource));
  t.true(Condition.create('@type', '=', 'full').evaluate(resource));
  t.false(Condition.create('@id', '=', '321').evaluate(resource));

  t.true(Condition.create('@id', '!=', 321).evaluate(resource));
  t.true(Condition.create('@id', '!=', '321').evaluate(resource));
  t.false(Condition.create('@type', '!=', 'full').evaluate(resource));

  t.true(Condition.create('@type', '~=', 'f*').evaluate(resource));
  t.false(Condition.create('@type', '~=', '*r').evaluate(resource));
  t.true(Condition.create('@type', '~=', 'FULL').evaluate(resource)); // case insensitivity

  t.true(Condition.create('@user_id', '$in', ['432', 867]).evaluate(resource));
  t.true(Condition.create('@user_id', '$in', [432, '867']).evaluate(resource));
  t.false(Condition.create('@user_id', '$in', [432, 987]).evaluate(resource));

  t.false(Condition.create('@user_id', '$nin', ['867', '233']).evaluate(resource));
  t.true(Condition.create('@user_id', '$nin', [925, 222, 999]).evaluate(resource));

  t.true(Condition.create('@lol', '=', '\\@wut').evaluate(resource)); // eslint-disable-line no-useless-escape

  t.false(Condition.create('@kind', '~', '^pretty$').evaluate(resource));
  t.true(Condition.create('@kind', '~', '^[A-Z]retty$').evaluate(resource));

  t.true(Condition.create('@kind', '~*', '^pretty$').evaluate(resource));
  t.false(Condition.create('@kind', '!~*', '^pretty$').evaluate(resource));

  t.true(Condition.create('@kind', '!~', '^Ugly$').evaluate(resource));
});

test('condition toJSON', t => {
  t.is(JSON.stringify(Condition.create('@user_id', '=', '333')), '["@user_id","=","333"]');
  t.is(JSON.stringify(Condition.create('@user_id', '!=', 'null')), '["@user_id","!=","null"]');
});
