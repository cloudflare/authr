'use strict';

import test from 'ava';
import Permission from '../src/Permission';
import SlugSet from '../src/Permission/SlugSet';
import ConditionSet from '../src/Permission/ConditionSet';

test('normal permission construction', t => {
  var p = Permission.allow({
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

  t.is(p.toString(), '{"allow":{"rsrc_type":"zone","rsrc_match":[["@id","=","123"]],"action":"enabled_service_mode"}}');
});

test('undefined resource type throws error', t => {
  var p = Permission.allow({
    rsrc_match: [['@id', '=', '123']],
    action: 'enabled_service_mode'
  });
  t.throws(() => {
    p.resourceTypes();
  });
});

test('undefined resource match throws err', t => {
  var p = Permission.deny({
    rsrc_type: 'zone',
    action: 'enabled_service_mode'
  });
  t.throws(() => {
    p.conditions();
  });
});

test('undefined action throws error', t => {
  var p = Permission.allow({
    rsrc_type: 'zone',
    rsrc_match: [['@id', '=', '123']]
  });
  t.throws(() => {
    p.actions();
  });
});

test('denying permissions', t => {
  let sub = {
    [Symbol.for('permission.subject_permissions')]: () => [
      Permission.deny({
        rsrc_type: 'zone',
        rsrc_match: [['@id', '=', '254']],
        action: 'hack'
      }),
      Permission.allow('all')
    ]
  };

  let rsrc = {
    [Symbol.for('permission.resource_type')]: 'zone',
    [Symbol.for('permission.resource_attr')]: key => {
      let attrs = {
        id: '254'
      };

      return attrs[key] === undefined ? null : attrs[key];
    }
  };

  let otherResource = {
    [Symbol.for('permission.resource_type')]: 'zone',
    [Symbol.for('permission.resource_attr')]: key => {
      let attrs = {
        id: '255'
      };

      return attrs[key] === undefined ? null : attrs[key];
    }
  };

  t.false(Permission.can(sub, 'hack', rsrc));
  t.true(Permission.can(sub, 'hack', otherResource));
});
