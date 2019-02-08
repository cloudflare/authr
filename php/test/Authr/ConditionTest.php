<?php

namespace Cloudflare\Test\Authr;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Resource;
use Cloudflare\Authr\Condition;
use Cloudflare\Authr\Exception\InvalidConditionOperator;

class ConditionTest extends TestCase
{
    protected $testResource = null;

    public function setUp(): void
    {
        $this->testResource = Resource::adhoc('thing', [
            'id' => '123',
            'type' => 'cool',
            'arr' => ['foo' => 1],
            'umm' => '@wut',
            'appearance' => 'Pretty'
        ]);
    }

    public function testUnknownOperator()
    {
        $this->expectException(InvalidConditionOperator::class);
        $a = new Condition('@id', '@>', '4');
    }

    public function testEvaluateDefaultOperator()
    {
        $a = new Condition('@id', '=', '123');
        $b = new Condition('not-cool', '=', '@type');
        $this->assertTrue($a->evaluate($this->testResource));
        $this->assertFalse($b->evaluate($this->testResource));
    }

    private function newCondition($a, $b, $c)
    {
        return new Condition($a, $b, $c);
    }

    public function testEscapedValue()
    {
        $this->assertTrue($this->newCondition('@umm', '=', '\@wut')->evaluate($this->testResource));
    }

    public function testNullValue()
    {
        $this->assertTrue($this->newCondition('@idk', '=', null)->evaluate($this->testResource));
    }

    public function testRegExpConditions()
    {
        $this->assertTrue($this->newCondition('@appearance', '!~', '^pretty$')->evaluate($this->testResource));
        $this->assertFalse($this->newCondition('@appearance', '!~', '^Pretty$')->evaluate($this->testResource));

        $this->assertTrue($this->newCondition('@appearance', '~', '^Pre')->evaluate($this->testResource));
        $this->assertFalse($this->newCondition('@appearance', '~', '^P[0-9]e')->evaluate($this->testResource));

        $this->assertTrue($this->newCondition('@appearance', '~*', '^pretty')->evaluate($this->testResource));
        $this->assertFalse($this->newCondition('@appearance', '~*', '^ugly$')->evaluate($this->testResource));

        $this->assertTrue($this->newCondition('@appearance', '!~*', '^Ugly')->evaluate($this->testResource));
        $this->assertFalse($this->newCondition('@appearance', '!~*', '^pret')->evaluate($this->testResource));
    }
}
