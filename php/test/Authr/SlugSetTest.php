<?php

namespace Cloudflare\Test\Authr;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\SlugSet;

class SlugSetTest extends TestCase
{
    public function testWildcard()
    {
        $set = new SlugSet('*');
        $this->assertTrue($set->contains('foo'));
        $this->assertTrue($set->contains('bar'));
        $this->assertTrue($set->contains('anything_and_everything'));
    }

    public function testNormalSet()
    {
        $set = new SlugSet(['foo', 'bar']);
        $this->assertTrue($set->contains('foo'));
        $this->assertTrue($set->contains('bar'));
        $this->assertFalse($set->contains('thisthing'));
    }

    public function testBlacklistSet()
    {
        $set = new SlugSet([
            SlugSet::NOT => ['foo', 'bar'],
        ]);
        $this->assertFalse($set->contains('foo'));
        $this->assertFalse($set->contains('bar'));
        $this->assertTrue($set->contains('thisthing'));
    }

    public function testStringTransform()
    {
        $set = new SlugSet('foo');
        $this->assertTrue($set->contains('foo'));
        $this->assertFalse($set->contains('bar'));
        $this->assertFalse($set->contains('thisthing'));
    }

    public function testStringTransformBlacklist()
    {
        $set = new SlugSet([SlugSet::NOT => 'foo']);
        $this->assertFalse($set->contains('foo'));
        $this->assertTrue($set->contains('bar'));
        $this->assertTrue($set->contains('thisthing'));
    }

    /**
     * @expectedException Cloudflare\Authr\Exception
     */
    public function testConstructWeirdValue()
    {
        new SlugSet(111);
    }

    /**
     * @dataProvider provideJsonSerializeScenarios
     */
    public function testJsonSerialize(SlugSet $set, $expected)
    {
        $this->assertEquals($expected, json_encode($set));
    }

    public function provideJsonSerializeScenarios()
    {
        return [
            'normal set' => [
                new SlugSet(['foo', 'bar']),
                '["foo","bar"]',
            ],
            'blacklist set' => [
                new SlugSet([SlugSet::NOT => ['bar', 'foo']]),
                '{"$not":["bar","foo"]}',
            ],
            'wildcard set' => [
                new SlugSet('*'),
                '"*"',
            ],
            'single slug set' => [
                new SlugSet('foo'),
                '"foo"',
            ],
        ];
    }
}
