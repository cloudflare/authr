<?php

namespace Cloudflare\Test\Authr;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Resource;
use Cloudflare\Authr\ConditionSet;
use Cloudflare\Authr\Condition;
use Cloudflare\Authr\Exception\InvalidConditionSetException;

class ConditionSetTest extends TestCase
{
    protected $testResource = null;

    public function setUp(): void
    {
        $this->testResource = Resource::adhoc('thing', [
            'id' => '123',
            'type' => 'cool',
            'pop' => 'opo',
        ]);
    }

    public function testConstructWeirdValue()
    {
        $this->expectException(InvalidConditionSetException::class);
        $x = new ConditionSet(222);
    }

    public function testVacuousTruth()
    {
        $set = new ConditionSet([]);
        $this->assertTrue($set->evaluate($this->testResource));
    }

    public function testShortCircuit()
    {
        $resource = Resource::adhoc('thing', [
            'id' => '123',
            'type' => 'cool',
            'sc_test' => function () {
                $this->fail('ConditionSet did not short circuit evaluation');

                return 'foo';
            }
        ]);
        $set = new ConditionSet([
            ConditionSet::LOGICAL_OR => [
                ['@id', '=', '123'],
                ['@sc_test', '=', 'foo'],
            ],
        ]);
        $this->assertTrue($set->evaluate($resource));
    }

    /** @dataProvider provideTestEvaluateScenarios */
    public function testEvaluate($result, $setPlain)
    {
        $set = new ConditionSet($setPlain);
        $this->assertTrue($result === $set->evaluate($this->testResource));        
    }

    public function provideTestEvaluateScenarios()
    {
        return [
            [false, [
                ['@id', '=', '123'],
                ['@pop', '=', 'p0p'],
            ]],
            // test nested sets
            [true, [ // (id = 321 OR (type = cool AND pop = opo))
                ConditionSet::LOGICAL_OR => [
                    ['@id', '=', '321'],
                    [ConditionSet::LOGICAL_AND => [
                        ['@type', '=', 'cool'],
                        ['@pop', '=', 'opo']
                    ]]
                ]
            ]]
        ];
    }

    /**
     * @dataProvider provideTestJsonSerializeScenarios
     */
    public function testJsonSerialize($expected, $setRaw)
    {
        $this->assertEquals($expected, json_encode(new ConditionSet($setRaw)));
    }

    public function provideTestJsonSerializeScenarios()
    {
        return [
            'normal set' => [
                '[["@id","=","321"],["@type","=","cool"]]',
                [
                    ['@id', '=', '321'],
                    ['@type', '=', 'cool'],
                ],
            ],
            'nested sets' => [
                '[["@id","=","321"],{"$or":[["@type","=",null],["@pop","=","opo"]]},[["@id","=","555"],["@attr","~","foo*"]]]',
                [
                    ['@id', '=', '321'],
                    [ConditionSet::LOGICAL_OR => [
                        ['@type', '=', null],
                        ['@pop', '=', 'opo'],
                    ]],
                    [ConditionSet::LOGICAL_AND => [
                        ['@id', '=', '555'],
                        ['@attr', '~', 'foo*'],
                    ]],
                ],
            ],
            'one more for good luck' => [
                '{"$or":[["id","=","321"],[["type","=","cool"],["pop","=","opo"]]]}',
                [
                    ConditionSet::LOGICAL_OR => [
                        ['id', '=', '321'],
                        [ConditionSet::LOGICAL_AND => [
                            ['type', '=', 'cool'],
                            ['pop', '=', 'opo'],
                        ]],
                    ],
                ]
            ]
        ];
    }
}
