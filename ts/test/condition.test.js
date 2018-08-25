'use strict';

import test from 'ava';
import Condition from '../build/condition';
import { GET_RESOURCE_TYPE, GET_RESOURCE_ATTRIBUTE } from '../build';

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
    [GET_RESOURCE_TYPE]: () => 'zone',
    [GET_RESOURCE_ATTRIBUTE]: key => {
      var attrs = {
        id: 123,
        type: 'full',
        kind: 'Pretty',
        user_id: '867',
        lol: '@wut',
        wut: '???',
        stuff: [1, '2', 3, 'foo', 'bar']
      };
      return attrs[key] || null;
    }
  };

  t.true(new Condition('@id', '=', '123').evaluate(resource));
  t.true(new Condition('@id', '=', 123).evaluate(resource));
  t.true(new Condition('@type', '=', 'full').evaluate(resource));
  t.false(new Condition('@id', '=', '321').evaluate(resource));

  t.true(new Condition('@id', '!=', 321).evaluate(resource));
  t.true(new Condition('@id', '!=', '321').evaluate(resource));
  t.false(new Condition('@type', '!=', 'full').evaluate(resource));

  t.true(new Condition('@type', '~=', 'f*').evaluate(resource));
  t.false(new Condition('@type', '~=', '*r').evaluate(resource));
  t.true(new Condition('@type', '~=', 'FULL').evaluate(resource)); // case insensitivity

  t.true(new Condition('@user_id', '$in', ['432', 867]).evaluate(resource));
  t.true(new Condition('@user_id', '$in', [432, '867']).evaluate(resource));
  t.false(new Condition('@user_id', '$in', [432, 987]).evaluate(resource));

  t.false(new Condition('@user_id', '$nin', ['867', '233']).evaluate(resource));
  t.true(new Condition('@user_id', '$nin', [925, 222, 999]).evaluate(resource));

  t.true(new Condition('@lol', '=', '\\@wut').evaluate(resource)); // eslint-disable-line no-useless-escape

  t.false(new Condition('@kind', '~', '^pretty$').evaluate(resource));
  t.true(new Condition('@kind', '~', '^[A-Z]retty$').evaluate(resource));

  t.true(new Condition('@kind', '~*', '^pretty$').evaluate(resource));
  t.false(new Condition('@kind', '!~*', '^pretty$').evaluate(resource));

  t.true(new Condition('@kind', '!~', '^Ugly$').evaluate(resource));

  t.true(new Condition('@stuff', '&', ['foo', 23]).evaluate(resource));
  t.false(new Condition('@stuff', '&', ['one', 'two']).evaluate(resource));
  t.true(new Condition('@stuff', '&', ['1']).evaluate(resource));
  t.true(new Condition('@stuff', '-', ['three', 'four']).evaluate(resource));
  t.false(new Condition('@stuff', '-', ['foo']).evaluate(resource));
  t.false(new Condition('@stuff', '-', ['3']).evaluate(resource));
});

test('condition toJSON', t => {
  t.is(JSON.stringify(new Condition('@user_id', '=', '333')), '["@user_id","=","333"]');
  t.is(JSON.stringify(new Condition('@user_id', '!=', 'null')), '["@user_id","!=","null"]');
});
