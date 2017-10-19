'use strict';

import test from 'ava';
import SlugSet from '../../src/Permission/SlugSet';

test('malformed slug set throws error', t => {
  t.throws(() => {
    SlugSet.create({ $lolnot: 'zone' });
  });
});

test('weird construction values throw errors', t => {
  t.throws(() => {
    SlugSet.create(null);
  });
  t.throws(() => {
    SlugSet.create({ $not: 5 });
  });
  t.throws(() => {
    SlugSet.create({});
  });
});

test('global matcher matches everything', t => {
  var set = SlugSet.create('*');
  t.true(set.contains('zone'));
  t.true(set.contains('record'));
  t.true(set.contains('user'));
});

test('single value slug set matches just one thing', t => {
  var set = SlugSet.create('zone');
  t.true(set.contains('zone'));
  t.false(set.contains('record'));
  t.false(set.contains('user'));
});

test('single value $not slug set matches everything else', t => {
  var set = SlugSet.create({ $not: 'zone' });
  t.false(set.contains('zone'));
  t.true(set.contains('record'));
  t.true(set.contains('user'));
});

test('multi-value slug set matches the strings it contains', t => {
  var set = SlugSet.create(['zone', 'record']);
  t.true(set.contains('zone'));
  t.true(set.contains('record'));
  t.false(set.contains('user'));
});

test('multi-value $not slug set matches the strings it does NOT contain', t => {
  var set = SlugSet.create({ $not: ['zone', 'record'] });
  t.false(set.contains('zone'));
  t.false(set.contains('record'));
  t.true(set.contains('user'));
});

test('SlugSet toString', t => {
  t.is(JSON.stringify(SlugSet.create('*')), '"*"');
  t.is(JSON.stringify(SlugSet.create(['zone'])), '"zone"');
  t.is(JSON.stringify(SlugSet.create(['zone', 'record'])), '["zone","record"]');
  t.is(JSON.stringify(SlugSet.create({ $not: ['zone'] })), '{"$not":"zone"}');
  t.is(JSON.stringify(SlugSet.create({ $not: ['zone', 'record'] })), '{"$not":["zone","record"]}');
});
