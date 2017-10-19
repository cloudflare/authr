<?php

namespace Cloudflare\Test\Authr\Condition\Operator\RegExp;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\RegExp\CaseInsensitive;

class CaseInsensitiveTest extends TestCase
{
    public function testCaseInsensitive()
    {
        $ci = new CaseInsensitive();
        $this->assertTrue($ci('FoOOBaarR', '^f{1}o+ba+r+$'));
        $this->assertTrue($ci('AAAAA', '^a{1,}$'));
        $this->assertFalse($ci('bbb', '^a+$'));
        $this->assertFalse($ci('CaseInsensitive', '^caseinsensitive--hi$'));
    }
}
