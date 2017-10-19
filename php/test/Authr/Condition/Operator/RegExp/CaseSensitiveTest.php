<?php

namespace Cloudflare\Test\Authr\Condition\Operator\RegExp;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\RegExp\CaseSensitive;

class CaseSensitiveTest extends TestCase
{
    public function testCaseSensitive()
    {
        $cs = new CaseSensitive();
        $this->assertFalse($cs('FoOOBaarR', '^f{1}o+ba+r+$'));
        $this->assertFalse($cs('AAAAA', '^a{1,}$'));
        $this->assertFalse($cs('bbb', '^a+$'));
        $this->assertFalse($cs('CaseSensitive', '^caseSensitive--hi$'));
        $this->assertTrue($cs('Hello There', '^(([A-Z][a-z]+)\s?)+$'));
        $this->assertTrue($cs('I capitalize my eyes', '\bI\b'));
        $this->assertTrue($cs('Remember Remember The Fifth of November!', 'Remember'));
    }
}
