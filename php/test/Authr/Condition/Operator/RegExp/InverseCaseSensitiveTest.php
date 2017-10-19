<?php

namespace Cloudflare\Test\Authr\Condition\Operator\RegExp;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Condition\Operator\RegExp\InverseCaseSensitive;

class InverseCaseSensitiveTest extends TestCase
{
    public function testInverseCaseSensitive()
    {
        $cs = new InverseCaseSensitive();
        $this->assertTrue($cs('FoOOBaarR', '^f{1}o+ba+r+$'));
        $this->assertTrue($cs('AAAAA', '^a{1,}$'));
        $this->assertTrue($cs('bbb', '^a+$'));
        $this->assertTrue($cs('CaseSensitive', '^caseSensitive--hi$'));
        $this->assertFalse($cs('Hello There', '^(([A-Z][a-z]+)\s?)+$'));
        $this->assertFalse($cs('I capitalize my eyes', '\bI\b'));
        $this->assertFalse($cs('Remember Remember The Fifth of November!', 'Remember'));
    }
}
