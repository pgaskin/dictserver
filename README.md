# dictserver

Dictionary library and API server based on data from the *Project Gutenberg EBook of Webster's Unabridged Dictionary*. The API responses (as shown below) will remain backwards/forwards-compatible. If compatibility is broken, it will be under a `/v2/` URL and be released as v2.x.x.

See the dictionary folder for usage as a Go library and the parsing code. The library can be used to create dictionaries from other sources. The on-disk format and library interface will remain backwards-compatible on a best-effort basis. If backwards-compatiblility is broken, it will break with an error message (rather than silently and cause other issues).

**Sample:** https://dict.geek1011.net/word/example

```json
{
    "status": "success",
    "result": {
        "word": "example",
        "info": "Ex*am\"ple, n.",
        "etymology": "[A later form for ensample, fr. L. exemplum, orig., what is taken out of a larger quantity, as a sample, from eximere to take out. See Exempt, and cf. Ensample, Sample.]",
        "meanings": [
            {
                "text": "One or a portion taken to show the character or quality of the whole; a sample; a specimen.",
                "referenced_words": null
            },
            {
                "text": "That which is to be followed or imitated as a model; a pattern or copy.",
                "example": "For I have given you an example, that ye should do as John xiii. 15. I gave, thou sayest, the example; I led the way. Milton.",
                "referenced_words": null
            },
            {
                "text": "That which resembles or corresponds with something else; a precedent; a model.",
                "example": "Such temperate order in so fierce a cause Doth want example. Shak.",
                "referenced_words": null
            },
            {
                "text": "That which is to be avoided; one selected for punishment and to serve as a warning; a warning.",
                "example": "Hang him; he'll be made an example. Shak. Now these things were our examples, to the intent that we should not lust after evil things, as they also lusted. 1 Cor. x. 6.",
                "referenced_words": null
            },
            {
                "text": "An instance serving for illustration of a rule or precept, especially a problem to be solved, or a case to be determined, as an exercise in the application of the rules of any study or branch of science; as, in trigonometry and grammar, the principles and rules are illustrated by examples.",
                "referenced_words": null
            }
        ],
        "notes": [
            "Precedent; case; instance.   Example, Instance. The discrimination to be made between these two words relates to cases in which we give \"instances\" or \"examples\" of things done. An instance denotes the single case then \"standing\" before us; if there be others like it, the word does not express this fact. On the contrary, an example is one of an entire class of like things, and should be a true representative or sample of that class. Hence, an example proves a rule or regular course of things; an instance simply points out what may be true only in the case presented. A man's life may be filled up with examples of the self- command and kindness which marked his character, and may present only a solitary instance of haste or severity. Hence, the word \"example\" should never be used to describe what stands singly and alone. We do, however, sometimes apply the word instance to what is really an example, because we are not thinking of the latter under this aspect, but solely as a case which \"stands before us.\" See Precedent."
        ],
        "credit": "Webster's Unabridged Dictionary (1913)",
        "additional_words": [
            {
                "word": "example",
                "alternates": [
                    "exampled",
                    "exampling"
                ],
                "info": " Ex*am\"ple, v. t. [imp. & p. p. Exampled; p. pr. & vb. n. Exampling.]",
                "meanings": [
                    {
                        "text": "To set an example for; to give a precedent for; to exemplify; to give an instance of; to instance. [Obs.] \"I may example my digression by some mighty precedent.\" Shak.",
                        "example": "Burke devoted himself to this duty with a fervid assiduity that has not often been exampled, and has never been surpassed. J. Morley.",
                        "referenced_words": null
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            }
        ],
        "referenced_words": null
    }
}
```

**Sample:** https://dict.geek1011.net/word/arch

```json
{
    "status": "success",
    "result": {
        "word": "arch",
        "info": "Arch, n.",
        "etymology": "[F. arche, fr. LL. arca, for arcus. See Arc.]",
        "meanings": [
            {
                "text": "(Geom.) Any part of a curved line.",
                "referenced_words": null
            },
            {
                "text": "(Arch.) (a) Usually a curved member made up of separate wedge-shaped solids, with the joints between them disposed in the direction of the radii of the curve; used to support the wall or other weight above an opening. In this sense arches are segmental, round (i. e., semicircular), or pointed.",
                "example": "(b) A flat arch is a member constructed of stones cut into wedges or other shapes so as to support each other without rising in a curve. Note: Scientifically considered, the arch is a means of spanning an opening by resolving vertical pressure into horizontal or diagonal thrust.",
                "referenced_words": null
            },
            {
                "text": "Any place covered by an arch; an archway; as, to pass into the arch of a bridge.",
                "referenced_words": null
            },
            {
                "text": "Any curvature in the form of an arch; as, the arch of the aorta. \"Colors of the showery arch.\" Milton. Triumphal arch, a monumental structure resembling an arched gateway, with one or more passages, erected to commemorate a triumph.",
                "referenced_words": null
            }
        ],
        "credit": "Webster's Unabridged Dictionary (1913)",
        "additional_words": [
            {
                "word": "arch",
                "alternates": [
                    "arched",
                    "arching"
                ],
                "info": " Arch, v. t. [imp. & p. p. Arched; p. pr. & vb. n. Arching.]",
                "meanings": [
                    {
                        "text": "To cover with an arch or arches.",
                        "referenced_words": null
                    },
                    {
                        "text": "To form or bend into the shape of an arch. The horse arched his neck. Charlesworth.",
                        "referenced_words": null
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            },
            {
                "word": "arch",
                "info": " Arch, v. i.",
                "meanings": [
                    {
                        "text": "To form into an arch; to curve.",
                        "referenced_words": null
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            },
            {
                "word": "arch",
                "info": "Arch, a.",
                "etymology": "[See Arch-, pref.]",
                "meanings": [
                    {
                        "text": "Chief; eminent; greatest; principal. The most arch act of piteous massacre. Shak.",
                        "referenced_words": null
                    },
                    {
                        "text": "Cunning or sly; sportively mischievous; roguish; as, an arch look, word, lad.",
                        "example": "[He] spoke his request with so arch a leer. Tatler.",
                        "referenced_words": null
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            },
            {
                "word": "arch",
                "info": "Arch, n.",
                "etymology": "[See Arch-, pref.]",
                "meanings": [
                    {
                        "text": "A chief. [Obs.] My worthy arch and patron comes to-night. Shak.",
                        "referenced_words": null
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            },
            {
                "word": "arch",
                "info": "*arch.",
                "etymology": "[Gr. Arch, a.]",
                "meanings": [
                    {
                        "text": "A suffix meaning a ruler, as in monarch (a sole ruler).",
                        "referenced_words": null
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            }
        ],
        "referenced_words": [
            {
                "word": "arc",
                "info": "Arc, n.",
                "etymology": "[F. arc, L. arcus bow, arc. See Arch, n.]",
                "meanings": [
                    {
                        "text": "(Geom.) A portion of a curved line; as, the arc of a circle or of an ellipse.",
                        "referenced_words": null
                    },
                    {
                        "text": "A curvature in the shape of a circular arc or an arch; as, the colored arc (the rainbow); the arc of Hadley's quadrant.",
                        "referenced_words": null
                    },
                    {
                        "text": "An arch. [Obs.] Statues and trophies, and triumphal arcs. Milton.",
                        "referenced_words": null
                    },
                    {
                        "text": "The apparent arc described, above or below the horizon, by the sun or other celestial body. The diurnal arc is described during the daytime, the nocturnal arc during the night. Electric arc, Voltaic arc. See under Voltaic.",
                        "referenced_words": [
                            "voltaic"
                        ]
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            }
        ]
    }
}
```

**Sample:** https://dict.geek1011.net/word/where

```json
{
    "status": "success",
    "result": {
        "word": "where",
        "info": "Where, adv.",
        "etymology": "[OE. wher, whar, AS. hw; akin to D. waar, OS. hw, OHG. hwar, war, wa, G. wo, Icel. and Sw. hvar, Dan. hvor, Goth. hwar, and E. who; cf. Skr. karhi when. sq. root182. See Who, and cf. There.]",
        "meanings": [
            {
                "text": "At or in what place; hence, in what situation, position, or circumstances; -- used interrogatively.",
                "example": "God called unto Adam, . . . Where art thou Gen. iii. 9. Note: See the Note under What, pron., 1.",
                "referenced_words": null
            },
            {
                "text": "At or in which place; at the place in which; hence, in the case or instance in which; -- used relatively.",
                "example": "She visited that place where first she was so happy. Sir P. Sidney. Where I thought the remnant of mine age Should have been cherished by her childlike duty. Shak. Where one on his side fights, thousands will fly. Shak. But where he rode one mile, the dwarf ran four. Sir W. Scott.",
                "referenced_words": null
            },
            {
                "text": "To what or which place; hence, to what goal, result, or issue; whither; -- used interrogatively and relatively; as, where are you going But where does this tend Goldsmith.",
                "example": "Lodged in sunny cleft, Where the gold breezes come not. Bryant. Note: Where is often used pronominally with or without a preposition, in elliptical sentences for a place in which, the place in which, or what place. The star . . . stood over where the young child was. Matt. ii. 9. The Son of man hath not where to lay his head. Matt. viii. 20. Within about twenty paces of where we were. Goldsmith. Where did the minstrels come from Dickens. Note: Where is much used in composition with preposition, and then is equivalent to a pronoun. Cf. Whereat, Whereby, Wherefore, Wherein, etc. Where away (Naut.), in what direction; as, where away is the land",
                "referenced_words": null
            }
        ],
        "notes": [
            "See Whither."
        ],
        "credit": "Webster's Unabridged Dictionary (1913)",
        "additional_words": [
            {
                "word": "where",
                "info": " Where, conj.",
                "meanings": [
                    {
                        "text": "Whereas. And flight and die is death destroying death; Where fearing dying pays death servile breath. Shak.",
                        "referenced_words": null
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            },
            {
                "word": "where",
                "info": " Where, n.",
                "meanings": [
                    {
                        "text": "Place; situation. [Obs. or Colloq.] Finding the nymph asleep in secret where. Spenser.",
                        "referenced_words": null
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            },
            {
                "word": "wher",
                "alternates": [
                    "where"
                ],
                "info": "Wher, Where (, pron. & conj.",
                "etymology": "[See Whether.]",
                "meanings": [
                    {
                        "text": "Whether. [Sometimes written whe'r.] [Obs.] Piers Plowman. Men must enquire (this is mine assent), Wher she be wise or sober or dronkelewe. Chaucer.",
                        "referenced_words": null
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": [
                    "whether"
                ]
            },
            {
                "word": "whither",
                "alternates": [
                    "no whither",
                    "where",
                    "whither"
                ],
                "info": "Whith\"er, adv.",
                "etymology": "[OE. whider. AS. hwider; akin to E. where, who; cf. Goth. hvadre whither. See Who, and cf. Hither, Thither.]",
                "meanings": [
                    {
                        "text": "To what place; -- used interrogatively; as, whither goest thou \"Whider may I flee\" Chaucer.",
                        "example": "Sir Valentine, whither away so fast Shak.",
                        "referenced_words": null
                    },
                    {
                        "text": "To what or which place; -- used relatively. That no man should know . . . whither that he went. Chaucer. We came unto the land whither thou sentest us. Num. xiii. 27.",
                        "referenced_words": null
                    },
                    {
                        "text": "To what point, degree, end, conclusion, or design; whereunto; whereto; -- used in a sense not physical.",
                        "example": "Nor have I . . . whither to appeal. Milton. Any whither, to any place; anywhere. [Obs.] \"Any whither, in hope of life eternal.\" Jer. Taylor.",
                        "referenced_words": null
                    }
                ],
                "notes": [
                    "No whither, to no place; nowhere. [Obs.] 2 Kings v. 25.   Where.   Whither, Where. Whither properly implies motion to place, and where rest in a place. Whither is now, however, to a great extent, obsolete, except in poetry, or in compositions of a grave and serious character and in language where precision is required. Where has taken its place, as in the question, \"Where are you going\""
                ],
                "extra": " Syn.",
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            }
        ],
        "referenced_words": [
            {
                "word": "whether",
                "info": "Wheth\"er, pron.",
                "etymology": "[OE. whether, AS. hwæ; akin to OS. hwe, OFries. hweder, OHG. hwedar, wedar, G. weder, conj., neither, Icel. hvarr whether, Goth. hwa, Lith. katras, L. uter, Gr. katara, from the interrogatively pronoun, in AS. hwa who. Who, and cf. Either, Neither, Or, conj.]",
                "meanings": [
                    {
                        "text": "Which (of two); which one (of two); -- used interrogatively and relatively. [Archaic] Now choose yourself whether that you liketh. Chaucer.",
                        "example": "One day in doubt I cast for to compare Whether in beauties' glory did exceed. Spenser. Whether of them twain did the will of his father Matt. xxi. 31.",
                        "referenced_words": null
                    }
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            },
            {
                "word": "whether",
                "alternates": [
                    "whether that"
                ],
                "info": " Wheth\"er, conj.",
                "meanings": [
                    {
                        "text": "In case; if; -- used to introduce the first or two or more alternative clauses, the other or others being connected by or, or by or whether. When the second of two alternatives is the simple negative of the first it is sometimes only indicated by the particle not or no after the correlative, and sometimes it is omitted entirely as being distinctly implied in the whether of the first. And now who knows But you, Lorenzo, whether I am yours Shak. You have said; but whether wisely or no, let the forest judge. Shak. For whether we live, we live unto the Lord; and whether we die, we die unto the Lord; whether we live therefore, or die, we are the Lord's. Rom. xiv. 8.",
                        "example": "But whether thus these things, or whether not; Whether the sun, predominant in heaven, Rise on the earth, or earth rise on the sun, . . . Solicit not thy thoughts with matters hid. Milton. Whether or no, in either case; in any case; as, I will go whether or no.",
                        "referenced_words": null
                    }
                ],
                "notes": [
                    "Whether that, whether. Shak."
                ],
                "credit": "Webster's Unabridged Dictionary (1913)",
                "referenced_words": null
            }
        ]
    }
}
```

**Sample** https://dict.geek1011.net/word/nonexistent-word

Will return a 404 error. The response body may change in the future (it will
either be an error or a success with `[]` as the result).
