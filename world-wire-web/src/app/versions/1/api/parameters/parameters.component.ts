import { Component, OnInit, Input } from '@angular/core';
import { Parameter } from 'swagger-schema-official';
import { SwaggerService } from '../../../shared/services/swagger.service';
import * as _ from 'lodash';

@Component({
  selector: '[app-parameters]',
  templateUrl: './parameters.component.html',
  styleUrls: ['./parameters.component.scss']
})
export class ParametersComponent implements OnInit {

  @Input() parametersDef: Parameter;

  constructor(
    public swaggerService: SwaggerService
  ) { }

  ngOnInit() {
    // console.log(this.parametersDef);
  }

}
