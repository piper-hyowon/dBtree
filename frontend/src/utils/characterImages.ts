import backupComplete from '../assets/images/character/backup-complete.png';
import dbError from '../assets/images/character/db-error.png';
import dbRunning from '../assets/images/character/db-running.png';
import dbPaused from '../assets/images/character/db-paused.png';
import dbStopped from '../assets/images/character/db-stopped.png';
import dbDeleting from '../assets/images/character/db-deleting.png';
import emptyState from '../assets/images/character/empty-state.png'
import highUsage from '../assets/images/character/high-usage.png';
import lowCredits from '../assets/images/character/low-credits.png';
import maintenance from '../assets/images/character/maintenance.png';
import richInCredits from '../assets/images/character/rich-credits.png';
import React from "react";
import LoadingLemonSVG from "../components/common/LoadingCharacterSVG";
import {InstanceStatus} from "../types/database.types";

type CharacterImage = string | React.ComponentType<any>;

export const characterImages = {
    backupComplete,
    error: dbError,
    loading: LoadingLemonSVG,
    provisioning: LoadingLemonSVG,
    running: dbRunning,
    paused: dbPaused,
    stopped: dbStopped,
    deleting: dbDeleting,
    highUsage,
    lowCredits,
    maintenance,
    richInCredits,
    default: emptyState
};

export const getCharacterByStatus = (status: InstanceStatus): string | React.ComponentType => {
    switch (status) {
        case 'running':
            return characterImages.running;
        case 'stopped':
            return characterImages.stopped;
        case 'paused':
            return characterImages.paused;
        case 'provisioning':
            return characterImages.provisioning;
        case 'deleting':
            return characterImages.deleting;
        case 'error':
            return characterImages.error;
        default:
            return characterImages.default;
    }
};

export const isImageComponent = (image: CharacterImage): image is React.ComponentType<any> => {
    return typeof image === 'function';
};