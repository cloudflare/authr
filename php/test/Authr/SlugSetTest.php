<?php

namespace Cloudflare\Test\Authr;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\SlugSet;
use Cloudflare\Authr\Exception;

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

    public function testBlocklistSet()
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

    public function testStringTransformBlocklist()
    {
        $set = new SlugSet([SlugSet::NOT => 'foo']);
        $this->assertFalse($set->contains('foo'));
        $this->assertTrue($set->contains('bar'));
        $this->assertTrue($set->contains('thisthing'));
    }

    public function testConstructWeirdValue()
    {
        $this->expectException(Exception::class);
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
            'blocklist set' => [
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
