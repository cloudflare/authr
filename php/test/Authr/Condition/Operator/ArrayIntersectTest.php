<?php

namespace Cloudflare\Test\Authr\Condition\Operator;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\ArrayIntersect;

class ArrayIntersectTest extends TestCase
{
    public function testArrayIntersect()
    {
        $ai = new ArrayIntersect();
        $this->assertTrue($ai(['foo', 'bar'], ['bar', 'baz']));
        $this->assertTrue($ai(
            ['key' => 'is', 'not' => 'important'],
            ['just' => 'values', 'thats' => 'important']
        ));
        $this->assertTrue($ai([5], ['5']));
        $this->assertFalse($ai(['one', 'two'], ['three', 'four']));
        $this->assertFalse($ai(5, [5])); // false returned on non-array input
        $this->assertFalse($ai(
            ['key' => 'v'],
            ['foo' => 'bar']
        ));
    }
}
