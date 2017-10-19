<?php

namespace Cloudflare\Test\Authr\Condition\Operator\RegExp;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\RegExp\InverseCaseInsensitive;

class InverseCaseInsensitiveTest extends TestCase
{
    public function testInverseCaseInsensitive()
    {
        $ci = new InverseCaseInsensitive();
        $this->assertFalse($ci('FoOOBaarR', '^f{1}o+ba+r+$'));
        $this->assertFalse($ci('AAAAA', '^a{1,}$'));
        $this->assertTrue($ci('bbb', '^a+$'));
        $this->assertTrue($ci('CaseInsensitive', '^caseinsensitive--hi$'));
    }
}
