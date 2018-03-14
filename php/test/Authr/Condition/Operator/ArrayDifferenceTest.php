<?php

namespace Cloudflare\Test\Authr\Condition\Operator;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\ArrayDifference;

class ArrayDifferenceTest extends TestCase
{
    public function testArrayDifference()
    {
        $ai = new ArrayDifference();
        $this->assertFalse($ai(['foo', 'bar'], ['bar', 'baz']));
        $this->assertFalse($ai(
            ['key' => 'is', 'not' => 'important'],
            ['just' => 'values', 'thats' => 'important']
        ));
        $this->assertFalse($ai([5], ['5'])); // loose type equality
        $this->assertTrue($ai(['one', 'two'], ['three', 'four']));
        $this->assertFalse($ai(5, [5])); // false returned on non-array input
        $this->assertTrue($ai(
            ['key' => 'v'],
            ['foo' => 'bar']
        ));
    }
}
