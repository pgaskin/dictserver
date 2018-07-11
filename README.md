# dictserver
Dictionary library and API server based on data from the *Project Gutenberg EBook of Webster's Unabridged Dictionary*. See the dictionary folder for usage as a Go library and the parsing code.

**Sample:** https://dict.geek1011.net/word/example
````json
{
    "status": "successs",
    "result": {
        "word": "example",
        "info": "Ex*am\"ple, n. ",
        "etymology": "[A later form for ensample, fr. L. exemplum, orig., what is taken out of a larger quantity, as a sample, from eximere to take out. See Exempt, and cf. Ensample, Sample.] ",
        "meanings": [{
            "text": "One or a portion taken to show the character or quality of the whole; a sample; a specimen."
        }, {
            "text": "That which is to be followed or imitated as a model; a pattern or copy.",
            "example": "For I have given you an example, that ye should do as John xiii. 15. I gave, thou sayest, the example; I led the way. Milton."
        }, {
            "text": "That which resembles or corresponds with something else; a precedent; a model.",
            "example": "Such temperate order in so fierce a cause Doth want example. Shak."
        }, {
            "text": "That which is to be avoided; one selected for punishment and to serve as a warning; a warning.",
            "example": "Hang him; he'll be made an example. Shak. Now these things were our examples, to the intent that we should not lust after evil things, as they also lusted. 1 Cor. x. 6."
        }, {
            "text": "An instance serving for illustration of a rule or precept, especially a problem to be solved, or a case to be determined, as an exercise in the application of the rules of any study or branch of science; as, in trigonometry and grammar, the principles and rules are illustrated by examples.",
            "example": " Syn. -- Precedent; case; instance. -- Example, Instance. The discrimination to be made between these two words relates to cases in which we give \"instances\" or \"examples\" of things done. An instance denotes the single case then \"standing\" before us; if there be others like it, the word does not express this fact. On the contrary, an example is one of an entire class of like things, and should be a true representative or sample of that class. Hence, an example proves a rule or regular course of things; an instance simply points out what may be true only in the case presented. A man's life may be filled up with examples of the self- command and kindness which marked his character, and may present only a solitary instance of haste or severity. Hence, the word \"example\" should never be used to describe what stands singly and alone. We do, however, sometimes apply the word instance to what is really an example, because we are not thinking of the latter under this aspect, but solely as a case which \"stands before us.\" See Precedent."
        }],
        "credit": "Webster's Unabridged Dictionary (1913)"
    }
}
````