<?php

namespace Cloudflare\Test\Authr;

use Cloudflare\Test\TestCase;
use Cloudflare\Authr\Rule;

class RuleTest extends TestCase
{
    public function testAllow()
    {
        $rule = Rule::allow([
            Rule::RESOURCE_TYPE => 'post',
            Rule::RESOURCE_MATCH => [['@id', '=', '123']],
            Rule::ACTION => 'update'
        ]);

        $this->assertEquals(Rule::ALLOW, $rule->access());
        $this->assertTrue($rule->resourceTypes()->contains('post'));
        $this->assertFalse($rule->resourceTypes()->contains('user'));
        $this->assertTrue($rule->actions()->contains('update'));
        $this->assertFalse($rule->actions()->contains('delete'));
    }

    public function testDeny()
    {
        $rule = Rule::deny([
            Rule::RESOURCE_TYPE => 'post',
            Rule::RESOURCE_MATCH => [['@id', '=', '123']],
            Rule::ACTION => 'update'
        ]);

        $this->assertEquals(Rule::DENY, $rule->access());
        $this->assertTrue($rule->resourceTypes()->contains('post'));
        $this->assertFalse($rule->resourceTypes()->contains('user'));
        $this->assertTrue($rule->actions()->contains('update'));
        $this->assertFalse($rule->actions()->contains('delete'));        
    }

    /** @expectedException Cloudflare\Authr\Exception\RuntimeException */
    public function testJSONDecodeFail()
    {
        $rulejson = '{"access":"all'; // eek! bad json!
        $rule = Rule::create($rulejson);
    }

    public function testJSONDecode()
    {
        $rulejson = '{"access":"allow","where":{"rsrc_type":"post","rsrc_match":[["@id","=","123"]],"action":"update"},"$meta":{"rule_id":123}}';
        $rule = Rule::create($rulejson);

        $this->assertEquals(Rule::ALLOW, $rule->access());
        $this->assertEquals('post', $rule->resourceTypes()->jsonSerialize());
        $this->assertEquals('update', $rule->actions()->jsonSerialize());
        $this->assertEquals([['@id', '=', '123']], $rule->conditions()->jsonSerialize());
        $this->assertEquals(['rule_id' => 123], $rule->meta());
    }

    /**
     * @dataProvider provideInvalidRuleScenarios
     * @expectedException Cloudflare\Authr\Exception\InvalidRuleException
     */
    public function testInvalidRule($ruleraw)
    {
        $rule = Rule::create($ruleraw);
    }

    public function provideInvalidRuleScenarios()
    {
        return [
            'wrong type' => [55],
            'missing "access"' => [
                [
                    Rule::WHERE => [
                        Rule::RESOURCE_TYPE => 'post',
                        Rule::RESOURCE_MATCH => [['@id', '=', '123']],
                        Rule::ACTION => 'update'
                    ]
                ]
            ],
            'missing "where"' => [
                [
                    Rule::ACCESS => Rule::ALLOW
                ]
            ],
            'invalid "access"' => [
                [
                    Rule::ACCESS => 'nah',
                    Rule::WHERE => [
                        Rule::RESOURCE_TYPE => 'post',
                        Rule::RESOURCE_MATCH => [['@id', '=', '123']],
                        Rule::ACTION => 'update'
                    ]
                ]
            ],
            'missing where.rsrc_type' => [
                [
                    Rule::ACCESS => Rule::ALLOW,
                    Rule::WHERE => [
                        Rule::RESOURCE_MATCH => [['@id', '=', '123']],
                        Rule::ACTION => 'update'
                    ]
                ]
            ],
            'missing where.rsrc_match' => [
                [
                    Rule::ACCESS => Rule::ALLOW,
                    Rule::WHERE => [
                        Rule::RESOURCE_TYPE => 'post',
                        Rule::ACTION => 'update'
                    ]
                ]
            ],
            'missing where.action' => [
                [
                    Rule::ACCESS => Rule::ALLOW,
                    Rule::WHERE => [
                        Rule::RESOURCE_TYPE => 'post',
                        Rule::RESOURCE_MATCH => [['@id', '=', '123']],
                    ]
                ]
            ],
            'unknown where key' => [
                [
                    Rule::ACCESS => Rule::ALLOW,
                    Rule::WHERE => [
                        Rule::RESOURCE_TYPE => 'post',
                        Rule::RESOURCE_MATCH => [['@id', '=', '123']],
                        Rule::ACTION => 'update',
                        'lol' => 'wut'
                    ]
                ]
            ]
        ];
    }

    /** @dataProvider provideRuleJsonSerializeScenarios */
    public function testRuleJsonSerialize($expected, $in)
    {
        $this->assertEquals($expected, json_encode($in));
    }

    public function provideRuleJsonSerializeScenarios()
    {
        return [
            [
                '{"access":"allow","where":{"rsrc_type":"post","rsrc_match":[["@id","=","123"]],"action":"update"},"$meta":{"rule_id":123}}',
                Rule::allow([
                    Rule::RESOURCE_TYPE => 'post',
                    Rule::RESOURCE_MATCH => [['@id', '=', '123']],
                    Rule::ACTION => 'update'
                ], ['rule_id' => 123])
            ],
            [
                '{"access":"allow","where":{"rsrc_type":{"$not":"post"},"rsrc_match":{"$or":[["@id","=","123"],["@name","$in",["foo","bar"]],[["@post_type","!=","pinned"],["@author_id","=","223"]]]},"action":["update","delete"]},"$meta":{"rule_id":321}}',
                Rule::allow([
                    Rule::RESOURCE_TYPE => ['$not' => ['post']],
                    Rule::RESOURCE_MATCH => [
                        '$or' => [
                            ['@id', '=', '123'],
                            ['@name', '$in', ['foo', 'bar']],
                            [
                                '$and' => [
                                    ['@post_type', '!=', 'pinned'],
                                    ['@author_id', '=', '223']
                                ]
                            ]
                        ]
                    ],
                    Rule::ACTION => ['update', 'delete']
                ], ['rule_id' => 321])
            ],
            'no meta' => [
                '{"access":"allow","where":{"rsrc_type":"post","rsrc_match":[["@id","=","123"]],"action":"update"}}',
                Rule::allow([
                    Rule::RESOURCE_TYPE => 'post',
                    Rule::RESOURCE_MATCH => [['@id', '=', '123']],
                    Rule::ACTION => 'update'
                ])
            ],
        ];
    }
}
