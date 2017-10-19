<?php

namespace Cloudflare\Test\Authr\Condition\Operator;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\Equals;

class EqualsTest extends TestCase
{
    public function testEquals()
    {
        $eq = new Equals();
        $this->assertTrue($eq('1', '1'));
        $this->assertFalse($eq('1', '0'));
        $this->assertTrue($eq('1', 1)); // loose equality
        $this->assertTrue($eq('foo', 'foo'));
        $this->assertFalse($eq('foo', 'bar'));
    }
}
