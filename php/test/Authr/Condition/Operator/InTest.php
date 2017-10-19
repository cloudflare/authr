<?php

namespace Cloudflare\Test\Authr\Condition\Operator;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\In;

class InTest extends TestCase
{
    public function testEquals()
    {
        $in = new In();
        $this->assertTrue($in('foo', ['foo', 'bar']));
        $this->assertFalse($in('foo', ['bar', 'baz']));
        $this->assertTrue($in(1, ['1', '2'])); // testing loose equality
        $this->assertFalse($in(1, null)); // testing polymorphic arguments
    }
}
