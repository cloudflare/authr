'use strict';

import test from 'ava';
import {
  can,
  Rule,
  GET_RULES,
  GET_RESOURCE_TYPE,
  GET_RESOURCE_ATTRIBUTE
} from '../src/authr';
import SlugSet from '../src/authr/SlugSet';
import ConditionSet from '../src/authr/ConditionSet';

test('normal permission construction', t => {
  var p = Rule.allow({
    rsrc_type: 'zone',
    rsrc_match: [
      ['@id', '=', '123']
    ],
    action: 'enabled_service_mode'
  });

  t.true(p.resourceTypes() instanceof SlugSet);
  t.true(p.resourceTypes().contains('zone'));
  t.false(p.resourceTypes().contains('record'));

  t.true(p.actions() instanceof SlugSet);
  t.false(p.actions().contains('delete'));
  t.true(p.actions().contains('enabled_service_mode'));

  t.true(p.conditions() instanceof ConditionSet);

  t.is(p.toString(), '{"access":"allow","where":{"action":"enabled_service_mode","rsrc_type":"zone","rsrc_match":[["@id","=","123"]]}}');
});

test('undefined resource type throws error', t => {
  var p = Rule.allow({
    rsrc_match: [['@id', '=', '123']],
    action: 'enabled_service_mode'
  });
  t.throws(() => {
    p.resourceTypes();
  });
});

test('undefined resource match throws err', t => {
  var p = Rule.deny({
    rsrc_type: 'zone',
    action: 'enabled_service_mode'
  });
  t.throws(() => {
    p.conditions();
  });
});

test('undefined action throws error', t => {
  var p = Rule.allow({
    rsrc_type: 'zone',
    rsrc_match: [['@id', '=', '123']]
  });
  t.throws(() => {
    p.actions();
  });
});

test('denying permissions', t => {
  let sub = {
    [GET_RULES]: () => [
      Rule.deny({
        rsrc_type: 'zone',
        rsrc_match: [['@id', '=', '254']],
        action: 'hack'
      }),
      Rule.allow('all')
    ]
  };

  let rsrc = {
    [GET_RESOURCE_TYPE]: () => 'zone',
    [GET_RESOURCE_ATTRIBUTE]: key => {
      let attrs = {
        id: '254'
      };
      return attrs[key] || null;
    }
  };

  let otherResource = {
    [GET_RESOURCE_TYPE]: () => 'zone',
    [GET_RESOURCE_ATTRIBUTE]: key => {
      let attrs = {
        id: '255'
      };
      return attrs[key] || null;
    }
  };

  t.false(can(sub, 'hack', rsrc));
  t.true(can(sub, 'hack', otherResource));
});
