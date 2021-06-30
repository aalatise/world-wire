import { CheckboxOption } from '../../../shared/models/checkbox-option.model';

export class Filter {
    key?: string;
    logic = '';
    value: string | number;
}

export interface CheckboxGroup {
    name: string;
    options: CheckboxOption[];
}

export interface CheckboxGroupFilter {
    [name: string]: CheckboxGroup;
}
