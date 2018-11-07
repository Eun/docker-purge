#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "jq.h"
#include "jv.h"


static int IsValidFilter(const char *filter) {
    jq_state *jq = jq_init();
    if (jq == NULL) {
        return 0;
    }
    if (jq_compile(jq, filter) <= 0) {
        return 0;
    }
    jq_teardown(&jq);
    return 1;
}

#define UNKNOWN_ERROR -1
#define JV_IS_INVALID -2
#define JQ_INIT_FAILED -3
#define JQ_COMPILE_FAILED -4

static int MatchesFilter(const char *in, const char* filter) {
    int result = UNKNOWN_ERROR;
    jv input = jv_parse(in);
    if (!jv_is_valid(input)) {
        result = JV_IS_INVALID;
        goto end;
    }

    jq_state *jq = jq_init();
    if (jq == NULL) {
        result = JQ_INIT_FAILED;
        goto end;
    }

    
    if (!jq_compile(jq, filter)) {
        result = JQ_COMPILE_FAILED;
        goto end;
    }
    jq_start(jq, input, 0);
    
    jv part = jq_next(jq);


    if (jv_is_valid(part)) {
        result = jv_equal(part, jv_true());
    } else {
        result = 0;
    }
end:
    jq_teardown(&jq);
    return result;
}
