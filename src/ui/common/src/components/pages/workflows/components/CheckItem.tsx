import { faCircleExclamation, faTriangleExclamation, faCircleCheck, faQuestionCircle, faCircleXmark } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { Box, Typography } from "@mui/material";
import { useState } from "react";
import { CheckLevel } from "../../../../utils/operators";
import ExecutionStatus from "../../../../utils/shared";
import { theme } from "../../../../styles/theme/theme";

const errorIcon = (<Box sx={{ fontSize: '20px', color: theme.palette.red['500'] }}>
    <FontAwesomeIcon icon={faCircleExclamation} />
</Box>);

const warningIcon = (<Box sx={{ fontSize: '20px', color: theme.palette.orange['500'] }}>
    <FontAwesomeIcon icon={faTriangleExclamation} />
</Box>);

const successIcon = (<Box sx={{ fontSize: '20px', color: theme.palette.green['400'] }}>
    <FontAwesomeIcon icon={faCircleCheck} />
</Box>);

const unknownIcon = (
    <Box sx={{ fontSize: '20px', color: theme.palette.gray['400'] }}>
        <FontAwesomeIcon icon={faQuestionCircle} />
    </Box>
);

const canceledIcon = (
    <Box sx={{ fontSize: '20px', color: theme.palette.gray['400'] }}>
        <FontAwesomeIcon icon={faCircleXmark} />
    </Box>
);

interface CheckPreview {
    checkId: string;
    name: string;
    status: ExecutionStatus;
    level: CheckLevel;
    value?: string;
    // a date.toLocaleString() should go here.
    timestamp: string;
}

interface CheckItemProps {
    checks: CheckPreview[];
}

export const CheckItem: React.FC<CheckItemProps> = ({ checks }) => {
    const [expanded, setExpanded] = useState(false);
    const checksList = [];
    let checksToShow = checks.length;

    if (!expanded && checks.length > 3) {
        checksToShow = 3;
    }

    for (let i = 0; i < checksToShow; i++) {
        let statusIcon = successIcon;
        if (checks[i].status === ExecutionStatus.Failed) {
            statusIcon = errorIcon;
        } else if (checks[i].status === ExecutionStatus.Succeeded) {
            // now we check the value to see if we should show warning or error icon
            if (checks[i].value === 'False') {
                if (checks[i].level === CheckLevel.Error) {
                    statusIcon = errorIcon;
                } else {
                    statusIcon = warningIcon;
                }
            }
        } else if (checks[i].status === ExecutionStatus.Canceled) {
            statusIcon = canceledIcon;
        } else if (checks[i].status !== ExecutionStatus.Succeeded) {
            statusIcon = unknownIcon;
        }

        checksList.push(
            <Box display="flex" key={checks[i].checkId} justifyContent="space-between">
                <Typography variant="body1" sx={{ fontWeight: 'bold' }}>{checks[i].name}</Typography>
                {statusIcon}
            </Box>
        )
    }

    const toggleExpanded = () => {
        setExpanded(!expanded);
    };

    const showMoreStyles = {
        fontWeight: 'bold',
        color: theme.palette.gray['700'],
        cursor: 'pointer',
        '&:hover': { textDecoration: 'underline' },
    };

    // TODO: make into a component to share with checks/metrics list
    const showLess = (
        <Box>
            <Typography
                variant="body1"
                sx={showMoreStyles}
                onClick={toggleExpanded}
            >
                Show Less ...
            </Typography>
        </Box>
    );

    // TODO: make into a component to share with checks/metrics list
    const showMore = (
        <Box>
            <Typography
                variant="body1"
                sx={showMoreStyles}
                onClick={toggleExpanded}
            >
                Show More ...
            </Typography>
        </Box>
    );

    const showMoreComponent = expanded ? showLess : showMore;

    return (
        <Box>
            {checksList}
            {checks.length > 3 && showMoreComponent}
        </Box>
    );
}

export default CheckItem;
