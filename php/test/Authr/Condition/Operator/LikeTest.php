<?php

namespace Cloudflare\Test\Authr\Condition\Operator;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\Like;

class LikeTest extends TestCase
{
    public function testEquals()
    {
        $like = new Like();
        $this->assertTrue($like('foobar', 'f*'));
        $this->assertTrue($like('foobar', '*ba*'));
        $this->assertTrue($like('FOoBArRR', 'fooba*'));
        $this->assertFalse($like('barbaz', 'baz*'));
    }
}
