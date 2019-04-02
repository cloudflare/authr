<?php

namespace Cloudflare\Test;

use Cloudflare\Authr;
use Cloudflare\Authr\Resource;
use Cloudflare\Authr\Segment;
use Cloudflare\Authr\ConditionSet;
use Cloudflare\Test\Authr\TestSubject;
use Psr\Log\NullLogger;

class AuthrTest extends TestCase
{
    /**
     * @dataProvider provideTestCanScenarios
     */
    public function testCan($subjectRules, $ops)
    {
        $subject = new TestSubject();
        $subject->setRules(array_map(implode('::', [Authr\Rule::class, 'create']), $subjectRules));
        foreach ($ops as $op) {
            list($action, $resourceDefinition, $result) = $op;
            $resource = Resource::adhoc($resourceDefinition['type'], $resourceDefinition['attributes']);
            $this->assertTrue($result === (new Authr(new NullLogger()))->can($subject, $action, $resource));
        }
    }

    public function provideTestCanScenarios()
    {
        $pshort = function ($typ, $t, $m, $a) {
            return [
                Authr\Rule::ACCESS => $typ,
                Authr\Rule::WHERE => [
                    Authr\Rule::RESOURCE_TYPE => $t,
                    Authr\Rule::RESOURCE_MATCH => $m,
                    Authr\Rule::ACTION => $a
                ]
            ];
        };
        $rshort = function ($t, $a) {
            return [
                'type' => $t,
                'attributes' => $a,
            ];
        };

        return [
            'nominal scenario' => [
                [
                    $pshort(Authr\Rule::ALLOW, 'dohikee', [['@status', '=', 'useful']], 'prod'),
                    $pshort(Authr\Rule::ALLOW, 'thing', [['@status', '=', 'useless']], 'delete'),
                ],
                [['delete', $rshort('thing', ['status' => 'useless']), true]]
            ],

            'subject has no permissions, cannot do anything' => [
                [],
                [['delete', $rshort('zone', ['id' => '123', 'name' => 'example.com']), false]]
            ],

            'subject can manage a few records in a particular zone' => [
                [
                    $pshort(Authr\Rule::ALLOW, 'record', [['@zone_id', '=', '123'], ['@name', '$in', ['cdn.example.com', 'service.example.com']]], ['update', 'change_service_mode']),
                ],
                [['update', $rshort('record', ['name' => 'cdn.example.com', 'zone_id' => '123']), true]]
            ],

            'subject can manage a few records in a particular zone, but not this one' => [
                [
                    $pshort(Authr\Rule::ALLOW, 'record', [['zone_id', '=', '123'], ['name', '$in', ['cdn.example.com', 'service.example.com']]], ['update', 'change_service_mode']),
                ],
                [['update', $rshort('record', ['name' => 'blog.example.com', 'zone_id' => '123']), false]]
            ],

            'possible pitfall: blacklist ignored, action granted by lower-ranked permission' => [
                [
                    // permission evaluator will green-light resource match, then
                    // see "delete" in blacklist. returns false, continues to
                    // evaluate subsequent permission.
                    $pshort(Authr\Rule::ALLOW, 'record', [['@zone_id', '=', '123']], ['$not' => ['delete', 'change_service_mode']]),

                    // permission evaluator will green-light resource match, NOT see
                    // "delete" in its blacklist, return true. therefore, a
                    // permission that wanted to blacklist "delete" gets overridden
                    // by another permission
                    $pshort(Authr\Rule::ALLOW, 'record', [['@zone_id', '=', '123'], ['@type', '=', 'A']], ['$not' => 'change_service_mode'])
                ],
                [['delete', $rshort('record', ['zone_id' => '123', 'type' => 'A']), true]]
            ],

            'denying permission should explicitly deny something even though there are lower-ranked permissions that would potentially allow' => [
                [
                    $pshort(Authr\Rule::DENY, 'record', [['@zone_id', '=', '324']], 'delete'),
                    $pshort(Authr\Rule::ALLOW, 'record', [], '*') // allow any action on any record!
                ],
                [
                    ['delete', $rshort('record', ['zone_id' => '324', 'type' => 'AAAA']), false],
                    ['delete', $rshort('record', ['zone_id' => '325', 'type' => 'A']), true]
                ]
            ],

            'allow => all should allow everything' => [
                [[Authr\Rule::ACCESS => Authr\Rule::ALLOW, Authr\Rule::WHERE => 'all']],
                [
                    ['delete', $rshort('record', ['zone_id' => '324', 'type' => 'AAAA']), true],
                    ['delete', $rshort('record', ['zone_id' => '325', 'type' => 'A']), true],
                    ['destroy', $rshort('system', [['name' => 'system']]), true]
                ]
            ],

            'deny => all should reject everything' => [
                [[Authr\Rule::ACCESS => Authr\Rule::DENY, Authr\Rule::WHERE => 'all']],
                [
                    ['delete', $rshort('record', ['zone_id' => '324', 'type' => 'AAAA']), false],
                    ['delete', $rshort('record', ['zone_id' => '325', 'type' => 'A']), false],
                    ['destroy', $rshort('system', [['name' => 'system']]), false]
                ]
            ]
        ];
    }
}
