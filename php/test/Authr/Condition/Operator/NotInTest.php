<?php

namespace Cloudflare\Test\Authr\Condition\Operator;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\NotIn;

class NotInTest extends TestCase
{
    public function testEquals()
    {
        $nin = new NotIn();
        $this->assertFalse($nin('foo', ['foo', 'bar']));
        $this->assertTrue($nin('foo', ['bar', 'baz']));
        $this->assertFalse($nin(1, ['1', '2'])); // testing loose equality
        $this->assertFalse($nin(1, null)); // testing polymorphic arguments
    }
}
