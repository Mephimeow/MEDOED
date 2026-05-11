#include <iostream>
#include <libaudit.h>
#include <errno.h>
#include <string.h>

int main()
{
    int audit = audit_open();
    if (!audit)
    {
        std::cerr << "Траблы с libaudit: " << strerror(audit) << std::endl;
        return 1;
    }
    
}