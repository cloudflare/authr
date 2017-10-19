<?php

namespace Cloudflare\Test\Authr\Condition\Operator;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\NotEquals;

class NotEqualsTest extends TestCase
{
    public function testEquals()
    {
        $neq = new NotEquals();
        $this->assertFalse($neq('1', '1'));
        $this->assertTrue($neq('1', '0'));
        $this->assertFalse($neq('1', 1)); // loose equality
        $this->assertFalse($neq('foo', 'foo'));
        $this->assertTrue($neq('foo', 'bar'));
    }
}
