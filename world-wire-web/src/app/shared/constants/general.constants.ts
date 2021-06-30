import { environment } from '../../../environments/environment';

export const FI_TYPES = [
    'Money Transfer Operator',
    'Financial Institution',
    'Bank',
    'Central Bank',
    'Credit Union',
    'Other'
];

export const ROLES = {
    admin: {
        role: 'admin',
        text: 'Administrator'
    },
    manager: {
        role: 'manager',
        text: 'Manager'
    },
    viewer: {
        role: 'viewer',
        text: 'Viewer'
    }
};

/**
 * Holds list of Participant Node Environments
 * generated based on list the supported_envs
 */
export const ENVIRONMENT = environment.supported_env;

export const STATUS = [
    'pending', 'active', 'suspended'
];
