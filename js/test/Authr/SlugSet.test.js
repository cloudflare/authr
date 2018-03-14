'use strict';

import test from 'ava';
import SlugSet from '../../src/authr/SlugSet';

test('malformed slug set throws error', t => {
  t.throws(() => {
    new SlugSet({ $lolnot: 'zone' }); // eslint-disable-line no-new
  });
});

test('weird construction values throw errors', t => {
  t.throws(() => {
    new SlugSet(null); // eslint-disable-line no-new
  });
  t.throws(() => {
    new SlugSet({ $not: 5 }); // eslint-disable-line no-new
  });
  t.throws(() => {
    new SlugSet({}); // eslint-disable-line no-new
  });
});

test('global matcher matches everything', t => {
  var set = new SlugSet('*');
  t.true(set.contains('zone'));
  t.true(set.contains('record'));
  t.true(set.contains('user'));
});

test('single value slug set matches just one thing', t => {
  var set = new SlugSet('zone');
  t.true(set.contains('zone'));
  t.false(set.contains('record'));
  t.false(set.contains('user'));
});

test('single value $not slug set matches everything else', t => {
  var set = new SlugSet({ $not: 'zone' });
  t.false(set.contains('zone'));
  t.true(set.contains('record'));
  t.true(set.contains('user'));
});

test('multi-value slug set matches the strings it contains', t => {
  var set = new SlugSet(['zone', 'record']);
  t.true(set.contains('zone'));
  t.true(set.contains('record'));
  t.false(set.contains('user'));
});

test('multi-value $not slug set matches the strings it does NOT contain', t => {
  var set = new SlugSet({ $not: ['zone', 'record'] });
  t.false(set.contains('zone'));
  t.false(set.contains('record'));
  t.true(set.contains('user'));
});

test('SlugSet toString', t => {
  t.is(JSON.stringify(new SlugSet('*')), '"*"');
  t.is(JSON.stringify(new SlugSet(['zone'])), '"zone"');
  t.is(JSON.stringify(new SlugSet(['zone', 'record'])), '["zone","record"]');
  t.is(JSON.stringify(new SlugSet({ $not: ['zone'] })), '{"$not":"zone"}');
  t.is(JSON.stringify(new SlugSet({ $not: ['zone', 'record'] })), '{"$not":["zone","record"]}');
});
